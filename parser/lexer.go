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

// This lexer is structured like the lexer described in this talk:
// - Video: https://www.youtube.com/watch?v=HxaD_trXwRE
// - Slides: https://talks.golang.org/2011/lex.slide
// However it does not implement the "new API" on slide 39-42, because:
// - We don't need to run this lexer during initialization
// - Apparently the restriction was lifted in Go 1 anyway...

type lexeme struct {
	tok  token.Token
	text string
}

type stateFn func(*lexer) stateFn

type lexer struct {
	name    string
	input   string
	lexemes chan lexeme
	start   int
	pos     int
	width   int
}

func lex(name, input string) (*lexer, chan lexeme) {
	l := &lexer{name: name, input: input, lexemes: make(chan lexeme)}
	go l.run()
	return l, l.lexemes
}

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.lexemes)
}

func (l *lexer) emit(t token.Token) {
	l.emitWithText(t, l.input[l.start:l.pos])
}

func (l *lexer) emitWithText(t token.Token, text string) {
	l.lexemes <- lexeme{t, text}
	l.start = l.pos
}

const eof = -1

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	var r rune
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.lexemes <- lexeme{token.Illegal, fmt.Sprintf(format, args...)}
	return nil
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

func lexText(l *lexer) stateFn {
	l.acceptRun(whitespace)
	l.ignore()

	if r := l.peek(); r == eof {
		return nil
	}

	return lexString
}

func lexString(l *lexer) stateFn {
	// We keep track of the delimiter (if any) so that we know when this
	// string ends. But the lexer output does not include the opening, or
	// closing, quote runes, so ignore the first character if it is a quote.
	delimiter := whitespace
	if r := l.peek(); strings.ContainsRune(quotes, r) {
		delimiter = string(l.next())
		l.ignore()
	}

	// A string might contain escaped characters in it, such as "\\" (an
	// escaped backslash). The lexer replaces the escape sequences with the
	// escaped characters (i.e. "\\" with "\"). This is done by copying over
	// substrings from the input into a `string.Builder`, with escape
	// sequences replaced appropriately.
	var b strings.Builder

	for escape := false; ; {
		r := l.next()
		if escape {
			// For now, we just treat any escaped rune as a literal
			// of that rune (e.g. "\ " is an escaped space).
			// TODO: map escape sequences like "\n" into a newline.
			b.WriteRune(r)
			l.ignore()
			escape = false
			continue
		} else if r == '\\' {
			b.WriteString(l.save())
			escape = true
			continue
		} else if strings.ContainsRune(delimiter, r) {
			b.WriteString(l.save())
			l.emitWithText(token.String, b.String())
			return lexText
		} else if r == eof {
			b.WriteString(l.save())
			t := token.String
			if delimiter == `'` {
				t = token.SubString1
				b.WriteRune('\n')
			} else if delimiter == `"` {
				t = token.SubString2
				b.WriteRune('\n')
			}
			l.emitWithText(t, b.String())
			return lexText
		}
	}
}
