package bind

import (
	"fmt"

	"golang.org/x/tools/go/types"
)

// seqType returns a string that can be used for reading and writing a
// type using the seq library.
func seqType(t types.Type) string {
	if isErrorType(t) {
		return "UTF16"
	}
	switch t := t.(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.Int:
			return "Int"
		case types.Int8:
			return "Int8"
		case types.Int16:
			return "Int16"
		case types.Int32:
			return "Int32"
		case types.Int64:
			return "Int64"
		case types.Uint8:
			// TODO(crawshaw): questionable, but vital?
			return "Byte"
		// TODO(crawshaw): case types.Uint, types.Uint16, types.Uint32, types.Uint64:
		case types.Float32:
			return "Float32"
		case types.Float64:
			return "Float64"
		case types.String:
			return "UTF16"
		default:
			// Should be caught earlier in processing.
			panic(fmt.Sprintf("unsupported return type: %s", t))
		}
	case *types.Named:
		switch u := t.Underlying().(type) {
		case *types.Interface:
			return "Ref"
		default:
			panic(fmt.Sprintf("unsupported named seqType: %s / %T", u, u))
		}
	default:
		panic(fmt.Sprintf("unsupported seqType: %s / %T", t, t))
	}
}

func seqRead(o types.Type) string {
	t := seqType(o)
	return t + "()"
}

func seqWrite(o types.Type, name string) string {
	t := seqType(o)
	if t == "Ref" {
		// TODO(crawshaw): do something cleaner, i.e. genWrite.
		return t + "(" + name + ".ref())"
	}
	return t + "(" + name + ")"
}
