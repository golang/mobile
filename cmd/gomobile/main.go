// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:generate gomobile help documentation doc.go

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"unicode"
	"unicode/utf8"
)

func printUsage(w io.Writer) {
	bufw := bufio.NewWriter(w)
	if err := usageTmpl.Execute(bufw, commands); err != nil {
		panic(err)
	}
	bufw.Flush()
}

var gomobileName = "gomobile"

func main() {
	gomobileName = os.Args[0]
	flag.Usage = func() {
		printUsage(os.Stderr)
		os.Exit(2)
	}
	flag.Parse()
	log.SetFlags(0)

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
	}

	if args[0] == "help" {
		if len(args) == 3 && args[1] == "documentation" {
			helpDocumentation(args[2])
			return
		}
		help(args[1:])
		return
	}

	for _, cmd := range commands {
		if cmd.Name == args[0] {
			cmd.flag.Usage = func() {
				cmd.usage()
				os.Exit(1)
			}
			cmd.flag.Parse(args[1:])
			if err := cmd.run(cmd); err != nil {
				msg := err.Error()
				if msg != "" {
					fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
				}
				os.Exit(1)
			}
			return
		}
	}
	fmt.Fprintf(os.Stderr, "%s: unknown subcommand %q\nRun 'gomobile help' for usage.\n", os.Args[0], args[0])
	os.Exit(2)
}

func help(args []string) {
	if len(args) == 0 {
		printUsage(os.Stdout)
		return // succeeded at helping
	}
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: %s help command\n\nToo many arguments given.\n", gomobileName)
		os.Exit(2) // failed to help
	}

	arg := args[0]
	for _, cmd := range commands {
		if cmd.Name == arg {
			cmd.usage()
			return // succeeded at helping
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown help topic %#q.  Run '%s help'.\n", arg, gomobileName)
	os.Exit(2)
}

const documentationHeader = `// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// DO NOT EDIT. GENERATED BY 'gomobile help documentation doc.go'.
`

func helpDocumentation(path string) {
	w := new(bytes.Buffer)
	w.WriteString(documentationHeader)
	w.WriteString("\n/*\n")
	if err := usageTmpl.Execute(w, commands); err != nil {
		log.Fatal(err)
	}

	for _, cmd := range commands {
		r, rlen := utf8.DecodeRuneInString(cmd.Short)
		w.WriteString("\n\n")
		w.WriteRune(unicode.ToUpper(r))
		w.WriteString(cmd.Short[rlen:])
		w.WriteString("\n\nUsage:\n\n\tgomobile " + cmd.Name + " " + cmd.Usage + "\n")
		w.WriteString(cmd.Long)
	}

	w.WriteString("*/\npackage main\n")

	if err := ioutil.WriteFile(path, w.Bytes(), 0666); err != nil {
		log.Fatal(err)
	}
}

var commands = []*command{
	// TODO(crawshaw): cmdRun
	cmdBind,
	cmdBuild,
	cmdInit,
	cmdInstall,
	cmdVersion,
}

type command struct {
	run   func(*command) error
	flag  flag.FlagSet
	Name  string
	Usage string
	Short string
	Long  string
}

func (cmd *command) usage() {
	fmt.Fprintf(os.Stdout, "usage: %s %s %s\n%s", gomobileName, cmd.Name, cmd.Usage, cmd.Long)
}

var usageTmpl = template.Must(template.New("usage").Parse(
	`Gomobile is a tool for building and running mobile apps written in Go.

To install:

	$ go get golang.org/x/mobile/cmd/gomobile
	$ gomobile init

At least Go 1.5 is required. Until it is released, build tip from
source: http://golang.org/doc/install/source

Initialization rebuilds the standard library and may download
the Android NDK compiler.

Usage:

	gomobile command [arguments]

Commands:
{{range .}}
	{{.Name | printf "%-11s"}} {{.Short}}{{end}}

Use 'gomobile help [command]' for more information about that command.

NOTE: iOS support is not ready yet.
`))
