// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Gomobile is a tool for building and running mobile apps written in Go.

The tool is under development and not ready for use.
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
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

var commands = []*command{
	// TODO(crawshaw): cmdRun
	cmdBind,
	cmdBuild,
	cmdInit,
	cmdInstall,
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
	`Gomobile is a tool for building Android and iOS Go apps.

Usage:

	gomobile command [arguments]

Commands:
{{range .}}
	{{.Name | printf "%-11s"}} {{.Short}}{{end}}

Use 'gomobile help [command]' for more information about that command.
`))
