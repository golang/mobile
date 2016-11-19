// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The java package takes the result of an AST traversal by the
// importers package and queries the java command for the type
// information for the referenced Java classes and interfaces.
//
// It is the of go/types for Java types and is used by the bind
// package to generate Go wrappers for Java API on Android.
package java

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/mobile/internal/importers"
)

// Class is the bind representation of a Java class or
// interface.
// Use Import to convert class references to Class.
type Class struct {
	// "java.pkg.Class.Inner"
	Name string
	// "java.pkg.Class$Inner"
	FindName string
	// JNI mangled name
	JNIName string
	// "Inner"
	PkgName string
	Funcs   []*Func
	Methods []*Func
	// All methods, including methods from
	// supers.
	AllMethods []*Func
	Vars       []*Var
	Supers     []string
	Final      bool
	Abstract   bool
	Interface  bool
	Throwable  bool
	// Whether the class has a no-arg constructor
	HasNoArgCon bool
}

// Func is a Java static function or method or constructor.
type Func struct {
	FuncSig
	ArgDesc string
	GoName  string
	// Mangled JNI name
	JNIName     string
	Static      bool
	Abstract    bool
	Final       bool
	Public      bool
	Constructor bool
	Params      []*Type
	Ret         *Type
	Decl        string
	Throws      string
}

// FuncSig uniquely identifies a Java Func.
type FuncSig struct {
	Name string
	// The method descriptor, in JNI format.
	Desc string
}

// Var is a Java member variable.
type Var struct {
	Name   string
	Static bool
	Final  bool
	Val    string
	Type   *Type
}

// Type is a Java type.
type Type struct {
	Kind  TypeKind
	Class string
	Elem  *Type
}

type TypeKind int

type importer struct {
	bootclspath string
	clspath     string
	clsMap      map[string]*Class
}

const (
	Int TypeKind = iota
	Boolean
	Short
	Char
	Byte
	Long
	Float
	Double
	String
	Array
	Object
)

var (
	errClsNotFound = errors.New("class not found")
)

// IsAvailable returns whether the required tools are available for
// Import to work. In particular, IsAvailable checks the existence
// of the javap binary.
func IsAvailable() bool {
	_, err := javapPath()
	return err == nil
}

func javapPath() (string, error) {
	return exec.LookPath("javap")
}

// Import returns Java Class descriptors for a list of references.
//
// The javap command from the Java SDK is used to dump
// class information. Its output looks like this:
//
// Compiled from "System.java"
// public final class java.lang.System {
//   public static final java.io.InputStream in;
//     descriptor: Ljava/io/InputStream;
//   public static final java.io.PrintStream out;
//     descriptor: Ljava/io/PrintStream;
//   public static final java.io.PrintStream err;
//     descriptor: Ljava/io/PrintStream;
//   public static void setIn(java.io.InputStream);
//     descriptor: (Ljava/io/InputStream;)V
//
//   ...
//
// }
func Import(bootclasspath, classpath string, refs *importers.References) ([]*Class, error) {
	imp := &importer{
		bootclspath: bootclasspath,
		clspath:     classpath,
		clsMap:      make(map[string]*Class),
	}
	clsSet := make(map[string]struct{})
	var names []string
	for _, ref := range refs.Refs {
		// The reference could be to some/pkg.Class or some/pkg/Class.Identifier. Include both.
		pkg := strings.Replace(ref.Pkg, "/", ".", -1)
		for _, cls := range []string{pkg, pkg + "." + ref.Name} {
			if _, exists := clsSet[cls]; !exists {
				clsSet[cls] = struct{}{}
				names = append(names, cls)
			}
		}
	}
	classes, err := imp.importClasses(names, true)
	if err != nil {
		return nil, err
	}
	for _, cls := range classes {
		imp.fillAllMethods(cls)
	}
	imp.fillThrowables()
	for _, cls := range classes {
		imp.mangleOverloads(cls.AllMethods)
		imp.mangleOverloads(cls.Funcs)
	}
	imp.filterReferences(classes, refs)
	return classes, nil
}

func (v *Var) Constant() bool {
	return v.Static && v.Final && v.Val != ""
}

// Mangle a name according to
// http://docs.oracle.com/javase/6/docs/technotes/guides/jni/spec/design.html#wp16696
//
// TODO: Support unicode characters
func JNIMangle(s string) string {
	var m []byte
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '.', '/':
			m = append(m, '_')
		case '$':
			m = append(m, "_00024"...)
		case '_':
			m = append(m, "_1"...)
		case ';':
			m = append(m, "_2"...)
		case '[':
			m = append(m, "_3"...)
		default:
			m = append(m, c)
		}
	}
	return string(m)
}

func (c *Class) HasSuper() bool {
	return !c.Final && !c.Interface
}

func (t *Type) Type() string {
	switch t.Kind {
	case Int:
		return "int"
	case Boolean:
		return "boolean"
	case Short:
		return "short"
	case Char:
		return "char"
	case Byte:
		return "byte"
	case Long:
		return "long"
	case Float:
		return "float"
	case Double:
		return "double"
	case String:
		return "String"
	case Array:
		return t.Elem.Type() + "[]"
	case Object:
		return t.Class
	default:
		panic("invalid kind")
	}
}

func (t *Type) JNIType() string {
	switch t.Kind {
	case Int:
		return "jint"
	case Boolean:
		return "jboolean"
	case Short:
		return "jshort"
	case Char:
		return "jchar"
	case Byte:
		return "jbyte"
	case Long:
		return "jlong"
	case Float:
		return "jfloat"
	case Double:
		return "jdouble"
	case String:
		return "jstring"
	case Array:
		return "jarray"
	case Object:
		return "jobject"
	default:
		panic("invalid kind")
	}
}

func (t *Type) CType() string {
	switch t.Kind {
	case Int, Boolean, Short, Char, Byte, Long, Float, Double:
		return t.JNIType()
	case String:
		return "nstring"
	case Array:
		if t.Elem.Kind != Byte {
			panic("unsupported array type")
		}
		return "nbyteslice"
	case Object:
		return "jint"
	default:
		panic("invalid kind")
	}
}

func (t *Type) JNICallType() string {
	switch t.Kind {
	case Int:
		return "Int"
	case Boolean:
		return "Boolean"
	case Short:
		return "Short"
	case Char:
		return "Char"
	case Byte:
		return "Byte"
	case Long:
		return "Long"
	case Float:
		return "Float"
	case Double:
		return "Double"
	case String, Object, Array:
		return "Object"
	default:
		panic("invalid kind")
	}
}

func (j *importer) filterReferences(classes []*Class, refs *importers.References) {
	refFuncs := make(map[[2]string]struct{})
	for _, ref := range refs.Refs {
		pkgName := strings.Replace(ref.Pkg, "/", ".", -1)
		cls := j.clsMap[pkgName]
		if cls == nil {
			continue
		}
		refFuncs[[...]string{pkgName, ref.Name}] = struct{}{}
	}
	for _, cls := range classes {
		var filtered []*Func
		for _, f := range cls.Funcs {
			if _, exists := refFuncs[[...]string{cls.Name, f.GoName}]; exists {
				filtered = append(filtered, f)
			}
		}
		cls.Funcs = filtered
		filtered = nil
		for _, m := range cls.Methods {
			if _, exists := refs.Names[m.GoName]; exists {
				filtered = append(filtered, m)
			}
		}
		cls.Methods = filtered
		filtered = nil
		for _, m := range cls.AllMethods {
			if _, exists := refs.Names[m.GoName]; exists {
				filtered = append(filtered, m)
			}
		}
		cls.AllMethods = filtered
	}
}

func (j *importer) importClasses(names []string, allowMissingClasses bool) ([]*Class, error) {
	if len(names) == 0 {
		return nil, nil
	}
	args := []string{"-J-Duser.language=en", "-s", "-protected", "-constants"}
	if j.clspath != "" {
		args = append(args, "-classpath", j.clspath)
	}
	if j.bootclspath != "" {
		args = append(args, "-bootclasspath", j.bootclspath)
	}
	args = append(args, names...)
	javapPath, err := javapPath()
	if err != nil {
		return nil, err
	}
	javap := exec.Command(javapPath, args...)
	out, err := javap.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, fmt.Errorf("javap failed: %v: %s", err)
		}
		// Not every name is a Java class so an exit error from javap is not
		// fatal.
	}
	s := bufio.NewScanner(bytes.NewBuffer(out))
	var classes []*Class
	for _, name := range names {
		cls, err := j.scanClass(s, name)
		if err != nil {
			if allowMissingClasses && err == errClsNotFound {
				continue
			}
			return nil, err
		}
		classes = append(classes, cls)
		j.clsMap[name] = cls
	}
	// Include the methods from classes extended or implemented
	unkCls := classes
	for {
		var unknown []string
		for _, cls := range unkCls {
			unknown = j.unknownSuperClasses(cls, unknown)
		}
		if len(unknown) == 0 {
			break
		}
		newCls, err := j.importClasses(unknown, false)
		if err != nil {
			return nil, err
		}
		for _, cls := range newCls {
			j.clsMap[cls.Name] = cls
		}
		unkCls = newCls
	}
	return classes, nil
}

func (j *importer) unknownSuperClasses(cls *Class, unk []string) []string {
loop:
	for _, n := range cls.Supers {
		if s, exists := j.clsMap[n]; exists {
			unk = j.unknownSuperClasses(s, unk)
		} else {
			for _, u := range unk {
				if u == n {
					continue loop
				}
			}
			unk = append(unk, n)
		}
	}
	return unk
}

func (j *importer) scanClass(s *bufio.Scanner, name string) (*Class, error) {
	if !s.Scan() {
		return nil, fmt.Errorf("%s: missing javap header", name)
	}
	head := s.Text()
	if errPref := "Error: "; strings.HasPrefix(head, errPref) {
		msg := head[len(errPref):]
		if strings.HasPrefix(msg, "class not found: "+name) {
			return nil, errClsNotFound
		}
		return nil, errors.New(msg)
	}
	if !strings.HasPrefix(head, "Compiled from ") {
		return nil, fmt.Errorf("%s: unexpected header: %s", name, head)
	}
	if !s.Scan() {
		return nil, fmt.Errorf("%s: missing javap class declaration", name)
	}
	clsDecl := s.Text()
	cls, err := j.scanClassDecl(name, clsDecl)
	if err != nil {
		return nil, err
	}
	if len(cls.Supers) == 0 {
		if name == "java.lang.Object" {
			cls.HasNoArgCon = true
		} else if !cls.Interface {
			cls.Supers = append(cls.Supers, "java.lang.Object")
		}
	}
	cls.JNIName = JNIMangle(cls.Name)
	clsElems := strings.Split(cls.Name, ".")
	cls.PkgName = clsElems[len(clsElems)-1]
	var funcs []*Func
	for s.Scan() {
		decl := strings.TrimSpace(s.Text())
		if decl == "}" {
			break
		} else if decl == "" {
			continue
		}
		if !s.Scan() {
			return nil, fmt.Errorf("%s: missing descriptor for member %q", name, decl)
		}
		desc := strings.TrimSpace(s.Text())
		desc = strings.TrimPrefix(desc, "descriptor: ")
		var static, final, abstract, public bool
		// Trim modifiders from the declaration.
	loop:
		for {
			idx := strings.Index(decl, " ")
			if idx == -1 {
				break
			}
			keyword := decl[:idx]
			switch keyword {
			case "public":
				public = true
			case "protected", "native":
				// ignore
			case "static":
				static = true
			case "final":
				final = true
			case "abstract":
				abstract = true
			default:
				// Hopefully we reached the declaration now.
				break loop
			}
			decl = decl[idx+1:]
		}
		// Trim ending ;
		decl = decl[:len(decl)-1]
		if idx := strings.Index(decl, "("); idx != -1 {
			f, err := j.scanMethod(decl, desc, idx)
			if err != nil {
				return nil, fmt.Errorf("%s: %v", name, err)
			}
			if f != nil {
				f.Static = static
				f.Abstract = abstract
				f.Public = public || cls.Interface
				f.Final = final
				f.Constructor = f.Name == cls.FindName
				if f.Constructor {
					cls.HasNoArgCon = cls.HasNoArgCon || len(f.Params) == 0
					f.Public = f.Public && !cls.Abstract
					f.Name = "new"
					f.Ret = &Type{Class: name, Kind: Object}
				}
				funcs = append(funcs, f)
			}
		} else {
			// Member is a variable
			v, err := j.scanVar(decl, desc)
			if err != nil {
				return nil, fmt.Errorf("%s: %v", name, err)
			}
			if v != nil && public {
				v.Static = static
				v.Final = final
				cls.Vars = append(cls.Vars, v)
			}
		}
	}
	for _, f := range funcs {
		if f.Static || f.Constructor {
			cls.Funcs = append(cls.Funcs, f)
		} else {
			cls.Methods = append(cls.Methods, f)
		}
	}
	return cls, nil
}

func (j *importer) scanClassDecl(name string, decl string) (*Class, error) {
	cls := &Class{
		Name: name,
	}
	const (
		stMod = iota
		stName
		stExt
		stImpl
	)
	st := stMod
	var w []byte
	// if > 0, we're inside a generics declaration
	gennest := 0
	for i := 0; i < len(decl); i++ {
		c := decl[i]
		switch c {
		default:
			if gennest == 0 {
				w = append(w, c)
			}
		case '>':
			gennest--
		case '<':
			gennest++
		case '{':
			return cls, nil
		case ' ', ',':
			if gennest > 0 {
				break
			}
			switch w := string(w); w {
			default:
				switch st {
				case stName:
					if strings.Replace(w, "$", ".", -1) != strings.Replace(name, "$", ".", -1) {
						return nil, fmt.Errorf("unexpected name %q in class declaration: %q", w, decl)
					}
					cls.FindName = w
				case stExt:
					cls.Supers = append(cls.Supers, w)
				case stImpl:
					if !cls.Interface {
						cls.Supers = append(cls.Supers, w)
					}
				default:
					return nil, fmt.Errorf("unexpected %q in class declaration: %q", w, decl)
				}
			case "":
				// skip
			case "public":
				if st != stMod {
					return nil, fmt.Errorf("unexpected %q in class declaration: %q", w, decl)
				}
			case "abstract":
				if st != stMod {
					return nil, fmt.Errorf("unexpected %q in class declaration: %q", w, decl)
				}
				cls.Abstract = true
			case "final":
				if st != stMod {
					return nil, fmt.Errorf("unexpected %q in class declaration: %q", w, decl)
				}
				cls.Final = true
			case "interface":
				cls.Interface = true
				fallthrough
			case "class":
				if st != stMod {
					return nil, fmt.Errorf("unexpected %q in class declaration: %q", w, decl)
				}
				st = stName
			case "extends":
				if st != stName {
					return nil, fmt.Errorf("unexpected %q in class declaration: %q", w, decl)
				}
				st = stExt
			case "implements":
				if st != stName && st != stExt {
					return nil, fmt.Errorf("unexpected %q in class declaration: %q", w, decl)
				}
				st = stImpl
			}
			w = w[:0]
		}
	}
	return nil, fmt.Errorf("missing ending { in class declaration: %q", decl)
}

func (j *importer) scanVar(decl, desc string) (*Var, error) {
	v := new(Var)
	const eq = " = "
	idx := strings.Index(decl, eq)
	if idx != -1 {
		val, ok := j.parseJavaValue(decl[idx+len(eq):])
		if !ok {
			// Skip constants that cannot be represented in Go
			return nil, nil
		}
		v.Val = val
	} else {
		idx = len(decl)
	}
	for i := idx - 1; i >= 0; i-- {
		if i == 0 || decl[i-1] == ' ' {
			v.Name = decl[i:idx]
			break
		}
	}
	if v.Name == "" {
		return nil, fmt.Errorf("unable to parse member name from declaration: %q", decl)
	}
	typ, _, err := j.parseJavaType(desc)
	if err != nil {
		return nil, fmt.Errorf("invalid type signature for %s: %q", v.Name, desc)
	}
	v.Type = typ
	return v, nil
}

func (j *importer) scanMethod(decl, desc string, parenIdx int) (*Func, error) {
	// Member is a method
	f := new(Func)
	f.Desc = desc
	for i := parenIdx - 1; i >= 0; i-- {
		if i == 0 || decl[i-1] == ' ' {
			f.Name = decl[i:parenIdx]
			break
		}
	}
	if f.Name == "" {
		return nil, fmt.Errorf("unable to parse method name from declaration: %q", decl)
	}
	if desc[0] != '(' {
		return nil, fmt.Errorf("invalid descriptor for method %s: %q", f.Name, desc)
	}
	const throws = " throws "
	if idx := strings.Index(decl, throws); idx != -1 {
		f.Throws = decl[idx+len(throws):]
	}
	i := 1
	for desc[i] != ')' {
		typ, n, err := j.parseJavaType(desc[i:])
		if err != nil {
			return nil, fmt.Errorf("invalid descriptor for method %s: %v", f.Name, err)
		}
		i += n
		f.Params = append(f.Params, typ)
	}
	f.ArgDesc = desc[1:i]
	i++ // skip ending )
	if desc[i] != 'V' {
		typ, _, err := j.parseJavaType(desc[i:])
		if err != nil {
			return nil, fmt.Errorf("invalid descriptor for method %s: %v", f.Name, err)
		}
		f.Ret = typ
	}
	return f, nil
}

func (j *importer) fillThrowables() {
	thrCls, ok := j.clsMap["java.lang.Throwable"]
	if !ok {
		// If Throwable isn't in the class map
		// no imported class inherits from Throwable
		return
	}
	for _, cls := range j.clsMap {
		j.fillThrowableFor(cls, thrCls)
	}
}

func (j *importer) fillThrowableFor(cls, thrCls *Class) {
	if cls.Interface || cls.Throwable {
		return
	}
	cls.Throwable = cls == thrCls
	for _, name := range cls.Supers {
		sup := j.clsMap[name]
		j.fillThrowableFor(sup, thrCls)
		cls.Throwable = cls.Throwable || sup.Throwable
	}
}

func (j *importer) fillAllMethods(cls *Class) {
	if len(cls.AllMethods) > 0 {
		return
	}
	if len(cls.Supers) == 0 {
		cls.AllMethods = cls.Methods
		return
	}
	for _, supName := range cls.Supers {
		super := j.clsMap[supName]
		j.fillAllMethods(super)
	}
	methods := make(map[FuncSig]struct{})
	for _, supName := range cls.Supers {
		super := j.clsMap[supName]
		for _, f := range super.AllMethods {
			if _, exists := methods[f.FuncSig]; !exists {
				methods[f.FuncSig] = struct{}{}
				// Copy function so each class can have its own
				// JNI name mangling.
				cpf := *f
				cls.AllMethods = append(cls.AllMethods, &cpf)
			}
		}
	}
	for _, f := range cls.Methods {
		if _, exists := methods[f.FuncSig]; !exists {
			cls.AllMethods = append(cls.AllMethods, f)
		}
	}
}

// mangleOverloads assigns unique names to overloaded Java functions by appending
// the argument count. If multiple methods have the same name and argument count,
// the method signature is appended in JNI mangled form.
func (j *importer) mangleOverloads(allFuncs []*Func) {
	overloads := make(map[string][]*Func)
	for _, f := range allFuncs {
		overloads[f.Name] = append(overloads[f.Name], f)
	}
	for _, funcs := range overloads {
		for _, f := range funcs {
			f.GoName = initialUpper(f.Name)
			f.JNIName = JNIMangle(f.Name)
		}
		if len(funcs) == 1 {
			continue
		}
		lengths := make(map[int]int)
		for _, f := range funcs {
			f.JNIName += "__" + JNIMangle(f.ArgDesc)
			lengths[len(f.Params)]++
		}
		for _, f := range funcs {
			n := len(f.Params)
			if lengths[n] > 1 {
				f.GoName += "_" + JNIMangle(f.ArgDesc)
				continue
			}
			if n > 0 {
				f.GoName = fmt.Sprintf("%s%d", f.GoName, n)
			}
		}
	}
}

func (j *importer) parseJavaValue(v string) (string, bool) {
	v = strings.TrimRight(v, "df")
	switch v {
	case "", "NaN", "Infinity", "-Infinity":
		return "", false
	default:
		if v[0] == '\'' {
			// Skip character constants, since they can contain invalid code points
			// that are unacceptable to Go.
			return "", false
		}
		return v, true
	}
}

func (j *importer) parseJavaType(desc string) (*Type, int, error) {
	t := new(Type)
	var n int
	if desc == "" {
		return t, n, errors.New("empty type signature")
	}
	n++
	switch desc[0] {
	case 'Z':
		t.Kind = Boolean
	case 'B':
		t.Kind = Byte
	case 'C':
		t.Kind = Char
	case 'S':
		t.Kind = Short
	case 'I':
		t.Kind = Int
	case 'J':
		t.Kind = Long
	case 'F':
		t.Kind = Float
	case 'D':
		t.Kind = Double
	case 'L':
		var clsName string
		for i := n; i < len(desc); i++ {
			if desc[i] == ';' {
				clsName = strings.Replace(desc[n:i], "/", ".", -1)
				n += i - n + 1
				break
			}
		}
		if clsName == "" {
			return t, n, errors.New("missing ; in class type signature")
		}
		if clsName == "java.lang.String" {
			t.Kind = String
		} else {
			t.Kind = Object
			t.Class = clsName
		}
	case '[':
		et, n2, err := j.parseJavaType(desc[n:])
		if err != nil {
			return t, n, err
		}
		n += n2
		t.Kind = Array
		t.Elem = et
	default:
		return t, n, fmt.Errorf("invalid type signature: %s", desc)
	}
	return t, n, nil
}

func initialUpper(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}
