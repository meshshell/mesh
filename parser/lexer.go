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

func (l *lexeme) String() string {
	return fmt.Sprintf("%v(%v)", l.tok, l.text)
}

type stateFn func(*lexer) stateFn

type lexer struct {
	name    string
	lexemes chan lexeme
	state   stateFn
	input   string
	start   int
	pos     int
	width   int
}

func newLexer(name string) *lexer {
	return &lexer{name: name, lexemes: make(chan lexeme), state: lexStart}
}

func (l *lexer) lex(input string) {
	l.input = input
	l.start, l.pos, l.width = 0, 0, 0
	for l.pos < len(l.input) {
		l.state = l.state(l)
	}
	l.emit(token.Newline)
}

func (l *lexer) emit(t token.Token) {
	l.emitWithText(t, l.input[l.start:l.pos])
}

func (l *lexer) emitWithText(t token.Token, text string) {
	l.lexemes <- lexeme{t, text}
	l.start = l.pos
}

const endOfLine = -1

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return endOfLine
	}
	var r rune
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup(r rune) {
	if r == endOfLine {
		return
	}
	_, width := utf8.DecodeRune([]byte(string(r)))
	l.pos -= width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup(r)
	return r
}

func (l *lexer) accept(valid string) bool {
	r := l.next()
	if strings.ContainsRune(valid, r) {
		return true
	}
	l.backup(r)
	return false
}

func (l *lexer) acceptRun(valid string) {
	r := l.next()
	for strings.ContainsRune(valid, r) {
		r = l.next()
	}
	l.backup(r)
}

func (l *lexer) save() string {
	tmp := l.input[l.start : l.pos-l.width]
	l.ignore()
	return tmp
}

const digits = "0123456789"
const lowercase = "abcdefghijklmnopqrstuvwxyz"
const uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const special = ";|$<>"
const whitespace = " \t\n"
const quotes = `'"`

func lexStart(l *lexer) stateFn {
	l.acceptRun(whitespace)
	l.ignore()

	switch r := l.next(); r {
	case '\\':
		if r2 := l.peek(); r2 == endOfLine {
			l.ignore()
			return lexStart
		}
		return lexUnquoted
	case '\'':
		l.ignore()
		return lexSingleQuoted(l)
	case '"':
		l.ignore()
		return lexDoubleQuoted(l)
	case '|':
		l.emit(token.Pipe)
		return lexStart
	case endOfLine:
		return lexStart
	default:
		l.backup(r)
		return lexUnquoted
	}
}

func lexSingleQuoted(l *lexer) stateFn {
	return lexString(l, `'`, lexSingleQuoted)
}

func lexDoubleQuoted(l *lexer) stateFn {
	return lexString(l, `"`, lexDoubleQuoted)
}

func lexUnquoted(l *lexer) stateFn {
	return lexString(l, special+whitespace, lexUnquoted)
}

func lexString(l *lexer, delimiter string, continuation stateFn) stateFn {
	// A string might contain escaped characters in it, such as "\\" (an
	// escaped backslash). The lexer replaces the escape sequences with the
	// escaped characters (i.e. "\\" with "\"). This is done by copying over
	// substrings from the input into a `string.Builder`, with escape
	// sequences replaced appropriately.
	var b strings.Builder
	for escaped := false; ; {
		r := l.next()
		if escaped {
			// For now, we just treat any escaped rune as a literal
			// of that rune (e.g. "\ " is an escaped space).
			// TODO: map escape sequences like "\n" into a newline.
			if r != endOfLine {
				b.WriteRune(r)
				l.ignore()
			}
			escaped = false
			continue
		} else if r == '\\' {
			b.WriteString(l.save())
			escaped = true
			continue
		} else if strings.ContainsRune(delimiter, r) {
			b.WriteString(l.save())
			l.emitWithText(token.String, b.String())
			return lexStart
		} else if r == endOfLine {
			b.WriteString(l.save())
			if delimiter == `'` || delimiter == `"` {
				b.WriteRune('\n')
				l.emitWithText(token.SubString, b.String())
			} else {
				l.emitWithText(token.String, b.String())
			}
			return continuation
		}
	}
}
