// Copyright 2020 Sam Uong
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/meshshell/mesh/interpreter"
	"github.com/meshshell/mesh/parser"
)

type stdio struct {
	in  *os.File
	out io.Writer
	err io.Writer
}

func main() {
	std := &stdio{os.Stdin, os.Stdout, os.Stderr}
	os.Exit(mesh(os.Args[0], os.Args[1:], std))
}

func mesh(cmd string, args []string, std *stdio) int {
	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	fs.SetOutput(std.err)
	snippet := fs.String("c", "", "run command from argument string")
	if err := fs.Parse(args); err == flag.ErrHelp {
		return 0
	} else if err != nil {
		fmt.Fprintf(std.err, "mesh: %v\n", err)
		return 1
	}

	if *snippet != "" {
		s := newNonInteractive(strings.NewReader(*snippet))
		return repl("-c", s, std)
	} else if script := fs.Arg(0); script != "" {
		f, err := os.Open(script)
		if err != nil {
			fmt.Fprintf(std.err, "mesh: %v\n", err)
			return 1
		}
		defer f.Close()
		return repl(script, newNonInteractive(f), std)
	} else if !terminal.IsTerminal(int(std.in.Fd())) {
		return repl("(stdin)", newNonInteractive(std.in), std)
	} else {
		s, err := newInteractive()
		if err != nil {
			fmt.Fprintf(std.err, "mesh: %v\n", err)
			return 1
		}
		defer s.close_()
		return repl("(stdin)", s, std)
	}
}

func repl(filename string, s scanner, std *stdio) int {
	status := 0
	parse := parser.NewParser(filename)
	interp := &interpreter.Interpreter{
		Stdin:  std.in,
		Stdout: std.out,
		Stderr: std.err,
	}
	s.setPrompt("] ")
	for {
		line, err := s.readLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(std.err, "mesh: %v\n", err)
			continue
		}
		stmt, err := parse.Parse(line)
		if err != nil {
			status = 1
			fmt.Fprintf(std.err, "mesh: %v\n", err)
			continue
		}
		status, err = stmt.Visit(interp)
		if err != nil {
			if e, ok := err.(interpreter.ExitStatus); ok {
				status = int(e)
				break
			}
			status = 1
			fmt.Fprintf(std.err, "mesh: %v\n", err)
			continue
		}
	}
	return status
}
