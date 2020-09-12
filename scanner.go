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
	"errors"

	"github.com/chzyer/readline"
)

var errInterrupt = errors.New("scanner: interrupted")

type scanner interface {
	Err() error
	Scan() bool
	Text() string
}

type rlScanner struct {
	rl   *readline.Instance
	text string
	err  error
}

func newRLScanner() (*rlScanner, error) {
	rl, err := readline.New("] ")
	if err != nil {
		return nil, err
	}
	rl.SetVimMode(true)
	return &rlScanner{rl, "", nil}, nil
}

func (s *rlScanner) Err() error {
	return s.err
}

func (s *rlScanner) Scan() bool {
	s.text, s.err = s.rl.Readline()
	if s.err == readline.ErrInterrupt {
		s.err = errInterrupt
	}
	return s.err == nil
}

func (s *rlScanner) Text() string {
	return s.text
}
