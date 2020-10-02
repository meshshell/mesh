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
	"bufio"
	"errors"
	"io"

	"github.com/chzyer/readline"
)

var errIgnoreEOF = errors.New("use `exit` to leave the shell")

type scanner interface {
	readLine() (string, error)
	setIgnoreEOF(ignore bool)
	setPrompt(prompt string)
	setViMode(vi bool)
}

type interactive struct {
	r         *readline.Instance
	ignoreEOF bool
}

func newInteractive() (*interactive, error) {
	r, err := readline.New("")
	if err != nil {
		return nil, err
	}
	r.SetVimMode(true)
	return &interactive{r, true}, nil
}

func (i *interactive) close_() error {
	return i.r.Close()
}

func (i *interactive) readLine() (string, error) {
	line, err := i.r.Readline()
	if i.ignoreEOF && err == io.EOF {
		return line, errIgnoreEOF
	}
	return line, err
}

func (i *interactive) setIgnoreEOF(ignore bool) {
	i.ignoreEOF = true
}

func (i *interactive) setPrompt(prompt string) {
	i.r.SetPrompt(prompt)
}

func (i *interactive) setViMode(vi bool) {
	i.r.SetVimMode(vi)
}

type noninteractive struct {
	r io.Reader
	s *bufio.Scanner
}

func newNonInteractive(r io.Reader) *noninteractive {
	return &noninteractive{r, bufio.NewScanner(r)}
}

func (n *noninteractive) readLine() (string, error) {
	if n.s == nil {
		return "", io.EOF
	} else if !n.s.Scan() {
		err := n.s.Err()
		// According to <https://pkg.go.dev/bufio#Scanner>: "Scanning
		// stops unrecoverably at EOF, the first I/O error, or a token
		// too large to fit in the buffer". So as soon as we encounter
		// any of these situations, set `n.s` to `nil` and return
		// `io.EOF` on every subsequent call.
		n.s = nil
		if err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return n.s.Text(), nil
}

func (n *noninteractive) setIgnoreEOF(_ bool) {
	// We never want to ignore EOF in non-interactive mode, otherwise we'll
	// get stuck in an infinite loop when we hit the end of the script.
}

func (n *noninteractive) setPrompt(_ string) {
	// Do nothing.
}

func (n *noninteractive) setViMode(_ bool) {
	// Do nothing.
}
