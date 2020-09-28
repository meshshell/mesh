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
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/meshshell/mesh/token"
)

type lexeme struct {
	tok  token.Token
	text string
}

func (l lexeme) String() string {
	return fmt.Sprintf("%v(%q)", l.tok, l.text)
}

type stateFn func(*lexer, string, int) stateFn

type lexer struct {
	name    string
	lexemes chan lexeme
	state   stateFn
}

func newLexer(name string) *lexer {
	return &lexer{name: name, lexemes: make(chan lexeme), state: lexStart}
}

func (l *lexer) lex(line string) {
	l.state = l.state(l, line, 0)
}

const digits = "0123456789"
const lowercase = "abcdefghijklmnopqrstuvwxyz"
const uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const special = "|$"
const whitespace = " \t\n"
const quotes = `'"`

func lexStart(l *lexer, line string, pos int) stateFn {
	right := strings.TrimLeft(line, whitespace)
	left := line[0 : len(line)-len(right)]
	if left != "" {
		l.lexemes <- lexeme{token.Whitespace, left}
	}
	line = right
	pos += len(left)

	if line == "" {
		l.lexemes <- lexeme{token.Newline, line}
		return lexStart
	} else if line == "\\" {
		l.lexemes <- lexeme{token.EscapedNewline, line}
		return lexStart
	}

	switch r, width := utf8.DecodeRuneInString(line); r {
	case '$':
		l.lexemes <- lexeme{token.Dollar, string(r)}
		return lexIdentifier(l, line[width:], pos+width)
	case '|':
		l.lexemes <- lexeme{token.Pipe, string(r)}
		return lexStart(l, line[width:], pos+width)
	case '~':
		// TODO: extract an (optional) username, e.g. "~sam"
		l.lexemes <- lexeme{token.Tilde, string(r)}
		return lexStart(l, line[width:], pos+width)
	case '\'':
		return lexSingleQuoted(l, line[width:], pos+width)
	case '"':
		return lexDoubleQuoted(l, line[width:], pos+width)
	default:
		return lexUnquoted(l, line, pos)
	}
}

func lexIdentifier(l *lexer, line string, pos int) stateFn {
	r, size := utf8.DecodeRuneInString(line)
	if !strings.ContainsRune(lowercase+uppercase+"_", r) {
		return lexStart(l, line, pos)
	}
	index := strings.IndexFunc(line[size:], func(r rune) bool {
		return !strings.ContainsRune(digits+lowercase+uppercase+"_", r)
	})
	if index == -1 {
		// The identifier runs to the end of the line; let lexStart()
		// emit the newline token and finish up.
		l.lexemes <- lexeme{token.Identifier, line}
		return lexStart(l, "", pos+len(line))
	}
	l.lexemes <- lexeme{token.Identifier, line[0 : size+index]}
	return lexStart(l, line[size+index:], pos+size+index)
}

func lexSingleQuoted(l *lexer, line string, pos int) stateFn {
	return quoted(l, line, pos, '\'', lexSingleQuoted)
}

func lexDoubleQuoted(l *lexer, line string, pos int) stateFn {
	return quoted(l, line, pos, '"', lexDoubleQuoted)
}

func quoted(l *lexer, line string, pos int, quote rune, next stateFn) stateFn {
	text, size := decodeString(line, pos, string(quote))
	line = line[size:]
	pos += size
	if r, _ := utf8.DecodeRuneInString(line); r != quote {
		l.lexemes <- lexeme{token.SubString, text}
		l.lexemes <- lexeme{token.Newline, line}
		return next
	}
	l.lexemes <- lexeme{token.String, text}
	return lexStart(l, line[1:], pos+1)
}

func lexUnquoted(l *lexer, line string, pos int) stateFn {
	text, size := decodeString(line, pos, special+whitespace)
	line = line[size:]
	pos += size
	if line == "\\" {
		l.lexemes <- lexeme{token.SubString, text}
		l.lexemes <- lexeme{token.Newline, line}
		return lexUnquoted
	}
	l.lexemes <- lexeme{token.String, text}
	return lexStart(l, line, pos)
}

func decodeString(line string, pos int, delimiter string) (string, int) {
	escaped := false
	start := 0
	var text strings.Builder
	for i, r := range line {
		if escaped {
			escaped = false
			start = i + utf8.RuneLen(r)
			// For now, we just treat any escaped rune as a literal
			// of that rune (e.g. "\ " is an escaped space).
			// TODO: map escape sequences like "\n" into a newline.
			text.WriteRune(r)
			continue
		} else if r == '\\' {
			escaped = true
			text.WriteString(line[start:i])
			continue
		} else if strings.ContainsRune(delimiter, r) {
			text.WriteString(line[start:i])
			return text.String(), i
		}
	}
	if escaped {
		return text.String(), len(line) - 1
	}
	text.WriteString(line[start:])
	if delimiter == `'` || delimiter == `"` {
		text.WriteRune('\n')
	}
	return text.String(), len(line)
}
