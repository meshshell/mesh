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

package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/meshshell/mesh/token"
)

func TestLexemeString(t *testing.T) {
	l := &lexeme{token.SubString, "mesh"}
	assert.Equal(t, "SubString(mesh)", l.String())
}

func timeout(t *testing.T, d time.Duration, done chan struct{}) {
	select {
	case <-time.After(d):
		t.Fail()
	case <-done:
	}
}

func TestLexerStrings(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output []string
	}{
		{"Command", "ls -l", []string{"ls", "-l"}},
		{"ExtraSpaces", ` a  b\ c  `, []string{"a", "b c"}},
		{"SingleQuoted", `a 'b  c\'"'`, []string{"a", `b  c'"`}},
		{"DoubleQuoted", `a "b  c'\""`, []string{"a", `b  c'"`}},
		{"StartsWithEscape", "echo \\\\", []string{"echo", "\\" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := newLexer(t.Name())
			done := make(chan struct{})
			go func() {
				defer close(done)
				for _, want := range test.output {
					got := <-l.lexemes
					assert.Equal(t, token.String, got.tok)
					assert.Equal(t, want, got.text)
				}
				newline := <-l.lexemes
				assert.Equal(t, token.Newline, newline.tok)
			}()
			l.lex(test.input)
			timeout(t, 100*time.Millisecond, done)
		})
	}
}

func TestLexerMultiLineStrings(t *testing.T) {
	tests := []struct {
		name    string
		inputs  []string
		outputs []lexeme
	}{
		{
			"TwoLines",
			[]string{"echo 'two", "lines'"},
			[]lexeme{
				{token.String, "echo"},
				{token.SubString, "two\n"},
				{token.Newline, ""},
				{token.String, "lines"},
				{token.Newline, ""},
			},
		}, {
			"EscapedNewline",
			[]string{"echo \\", "foo"},
			[]lexeme{
				{token.String, "echo"},
				{token.Newline, ""},
				{token.String, "foo"},
				{token.Newline, ""},
			},
		}, {
			"StartsWithQuote",
			[]string{"'", "baz'"},
			[]lexeme{
				{token.SubString, "\n"},
				{token.Newline, ""},
				{token.String, "baz"},
				{token.Newline, ""},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := newLexer(test.name)
			done := make(chan struct{})
			go func() {
				defer close(done)
				for _, want := range test.outputs {
					got := <-l.lexemes
					assert.Equal(t, want, got,
						"wanted token.%v, got token.%v",
						want.tok, got.tok)
				}
			}()
			for _, line := range test.inputs {
				l.lex(line)
			}
			timeout(t, 100*time.Millisecond, done)
		})
	}
}
