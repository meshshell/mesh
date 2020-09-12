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
	"fmt"
	"io"
	"os"

	"github.com/chzyer/readline"

	"github.com/meshshell/mesh/parser"
)

func main() {
	rl, err := readline.New("] ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer rl.Close()
	rl.SetVimMode(true)
	for {
		line, err := rl.Readline()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		p := parser.NewParser("(stdin)")
		stmt, err := p.Parse(line)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if _, err = stmt.Exec(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
