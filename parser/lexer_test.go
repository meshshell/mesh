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
	l := lexeme{token.SubString, "mesh"}
	assert.Equal(t, `SubString("mesh")`, l.String())
}

type lexerTest struct {
	name    string
	inputs  []string
	outputs []lexeme
}

func (test *lexerTest) run(t *testing.T) {
	start := time.Now()
	lex := newLexer(test.name)
	lexerDone := make(chan struct{})
	go func() {
		defer close(lexerDone)
		for _, line := range test.inputs {
			lex.lex(line)
		}
	}()
	assertsDone := make(chan struct{})
	go func() {
		defer close(assertsDone)
		for _, want := range test.outputs {
			got := <-lex.lexemes
			assert.Equal(t, want, got, "want %v, got %v", want, got)
		}
	}()
	timeoutDuration := 100 * time.Millisecond
	timeout := time.After(timeoutDuration)
	select {
	case <-lexerDone:
		select {
		case <-assertsDone:
			break
		case <-timeout:
			t.Fatal("timed out waiting for asserts")
		}
	case <-assertsDone:
		select {
		case <-lexerDone:
			break
		case l := <-lex.lexemes:
			t.Fatalf("unexpected lexeme: %v", l)
		case <-timeout:
			t.Fatal("timed out waiting for lexer")
		}
	case <-timeout:
		t.Fatal("timed out waiting for lexer and asserts (deadlock?)")
	}
	if time.Since(start) >= timeoutDuration/2 {
		t.Logf("warning: %s took %v, consider increasing timeout",
			t.Name(), time.Since(start))
	}
}

func TestLexerStrings(t *testing.T) {
	for _, test := range []lexerTest{
		{
			"Command",
			[]string{"ls -l"},
			[]lexeme{
				{token.String, "ls"},
				{token.Whitespace, " "},
				{token.String, "-l"},
				{token.Newline, ""},
			},
		}, {
			"ExtraSpaces",
			[]string{` a  b\ c   `},
			[]lexeme{
				{token.Whitespace, " "},
				{token.String, "a"},
				{token.Whitespace, "  "},
				{token.String, "b c"},
				{token.Whitespace, "   "},
				{token.Newline, ""},
			},
		}, {
			"SingleQuoted",
			[]string{`a 'b  c\'"'`},
			[]lexeme{
				{token.String, "a"},
				{token.Whitespace, " "},
				{token.String, `b  c'"`},
				{token.Newline, ""},
			},
		}, {
			"DoubleQuoted",
			[]string{`a "b  c'\""`},
			[]lexeme{
				{token.String, "a"},
				{token.Whitespace, " "},
				{token.String, `b  c'"`},
				{token.Newline, ""},
			},
		}, {
			"StartsWithEscape",
			[]string{"echo \\\\"},
			[]lexeme{
				{token.String, "echo"},
				{token.Whitespace, " "},
				{token.String, "\\"},
				{token.Newline, ""},
			},
		},
	} {
		t.Run(test.name, test.run)
	}
}

func TestLexerMultiLineStrings(t *testing.T) {
	for _, test := range []lexerTest{
		{
			"QuotedOverTwoLines",
			[]string{"echo 'two", "lines'"},
			[]lexeme{
				{token.String, "echo"},
				{token.Whitespace, " "},
				{token.SubString, "two\n"},
				{token.Newline, ""},
				{token.String, "lines"},
				{token.Newline, ""},
			},
		}, {
			"UnquotedOverTwoLines",
			[]string{"echo two\\", "lines"},
			[]lexeme{
				{token.String, "echo"},
				{token.Whitespace, " "},
				{token.SubString, "two"},
				{token.Newline, "\\"},
				{token.String, "lines"},
				{token.Newline, ""},
			},
		}, {
			"EscapedNewline",
			[]string{"echo \\", "foo"},
			[]lexeme{
				{token.String, "echo"},
				{token.Whitespace, " "},
				{token.EscapedNewline, "\\"},
				{token.String, "foo"},
				{token.Newline, ""},
			},
		}, {
			"StartsWithQuote",
			[]string{"'", "bar'"},
			[]lexeme{
				{token.SubString, "\n"},
				{token.Newline, ""},
				{token.String, "bar"},
				{token.Newline, ""},
			},
		},
	} {
		t.Run(test.name, test.run)
	}
}

func TestLexerVariables(t *testing.T) {
	for _, test := range []lexerTest{
		{
			"OneLetterIdentifier",
			[]string{"cd $X"},
			[]lexeme{
				{token.String, "cd"},
				{token.Whitespace, " "},
				{token.Dollar, "$"},
				{token.Identifier, "X"},
				{token.Newline, ""},
			},
		}, {
			"StartOfWord",
			[]string{"cd $HOME"},
			[]lexeme{
				{token.String, "cd"},
				{token.Whitespace, " "},
				{token.Dollar, "$"},
				{token.Identifier, "HOME"},
				{token.Newline, ""},
			},
		}, {
			"MiddleOfWord",
			[]string{"cd /home/$USER/Desktop"},
			[]lexeme{
				{token.String, "cd"},
				{token.Whitespace, " "},
				{token.String, "/home/"},
				{token.Dollar, "$"},
				{token.Identifier, "USER"},
				{token.String, "/Desktop"},
				{token.Newline, ""},
			},
		}, {
			"EndOfWord",
			[]string{"cd X$"},
			[]lexeme{
				{token.String, "cd"},
				{token.Whitespace, " "},
				{token.String, "X"},
				{token.Dollar, "$"},
				{token.Newline, ""},
			},
		}, {
			"BeforeString",
			[]string{"cd $/X"},
			[]lexeme{
				{token.String, "cd"},
				{token.Whitespace, " "},
				{token.Dollar, "$"},
				{token.String, "/X"},
				{token.Newline, ""},
			},
		},
	} {
		t.Run(test.name, test.run)
	}
}

func TestLexerTildes(t *testing.T) {
	for _, test := range []lexerTest{
		{
			"Tilde",
			[]string{"cd ~"},
			[]lexeme{
				{token.String, "cd"},
				{token.Whitespace, " "},
				{token.Tilde, "~"},
				{token.Newline, ""},
			},
		}, {
			"TildeWithPath",
			[]string{"cd ~/bin"},
			[]lexeme{
				{token.String, "cd"},
				{token.Whitespace, " "},
				{token.Tilde, "~"},
				{token.String, "/bin"},
				{token.Newline, ""},
			},
		}, {
			// On the one hand, it would be nice to be able to write
			// `file://~/index.html` and have it expand to the
			// user's home directory. But on the other hand, `~` is
			// a traditional suffix for backup files (e.g., see
			// https://unix.stackexchange.com/q/76189). So, Mesh
			// follows Unix tradition here and only treats `~` as a
			// special character if it is at the start of a word.
			"TildeAtMiddleAndEndOfWordIsNotSpecial",
			[]string{"cd /~/~"},
			[]lexeme{
				{token.String, "cd"},
				{token.Whitespace, " "},
				{token.String, "/~/~"},
				{token.Newline, ""},
			},
		},
	} {
		t.Run(test.name, test.run)
	}
}

func TestLexerMultipleCommands(t *testing.T) {
	for _, test := range []lexerTest{
		{
			"Semicolon",
			[]string{"cd;ls"},
			[]lexeme{
				{token.String, "cd"},
				{token.Semicolon, ";"},
				{token.String, "ls"},
				{token.Newline, ""},
			},
		}, {
			"Pipeline",
			[]string{"sort|uniq"},
			[]lexeme{
				{token.String, "sort"},
				{token.Pipe, "|"},
				{token.String, "uniq"},
				{token.Newline, ""},
			},
		},
	} {
		t.Run(test.name, test.run)
	}
}
