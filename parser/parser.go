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
	"sync"

	"github.com/meshshell/mesh/ast"
	"github.com/meshshell/mesh/token"
)

type parserError struct {
	msg string
}

func (pe parserError) Error() string {
	return "parser: " + pe.msg
}

type Parser struct {
	lex       *lexer
	done      chan bool
	lock      sync.Mutex
	locked    bool
	stmt      ast.Stmt
	err       error
	lookahead *lexeme
}

func NewParser(filename string) *Parser {
	return &Parser{
		lex:  newLexer(filename),
		done: make(chan bool),
	}
}

func (p *Parser) Parse(line string) bool {
	if !p.locked {
		go p.parseStmt()
	}
	p.lex.lex(line)
	return <-p.done
}

func (p *Parser) Result() (ast.Stmt, error) {
	if p.locked {
		panic("parser: Parser.Result() called before parsing completed")
	}
	return p.stmt, p.err
}

func (p *Parser) next() *lexeme {
	if p.lookahead != nil {
		defer func() { p.lookahead = nil }()
		return p.lookahead
	}
	l := <-p.lex.lexemes
	return &l
}

func (p *Parser) peek() *lexeme {
	if p.lookahead == nil {
		l := <-p.lex.lexemes
		p.lookahead = &l
	}
	return p.lookahead
}

func (p *Parser) parseStmt() {
	p.lock.Lock()
	p.locked = true
	p.stmt, p.err = nil, nil
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(parserError); ok {
				p.err = err
				for p.next().tok != token.Newline {
					// If the parser panics before parsing
					// the current line, the lexer will
					// still continue to run. So we need to
					// drain the p.lexemes channel of all
					// tokens until the next Newline, so
					// that the lexer doesn't block.
				}
			} else {
				panic(r)
			}
		}
		p.locked = false
		p.done <- true
		p.lock.Unlock()
	}()

	var cmd *ast.Cmd = &ast.Cmd{}
	for l := p.peek(); l.tok == token.Whitespace; l = p.next() {
		// Trim any leading whitespace.
	}
	switch l := p.peek(); l.tok {
	case token.Newline:
		break
	case token.Pipe:
		panic(parserError{"pipe operator not yet implemented"})
	default:
		cmd = p.parseCmd()
	}
	switch l := p.next(); l.tok {
	case token.Newline:
		p.stmt = cmd
	default:
		p.err = fmt.Errorf("parser: unexpected token: %v", l)
	}
}

func (p *Parser) parseCmd() *ast.Cmd {
	var argv []ast.Expr
	for {
		switch l := p.peek(); l.tok {
		case token.EscapedNewline:
			p.done <- false
			p.next()
			continue
		case token.Whitespace:
			p.next()
			continue
		case token.String, token.SubString, token.Tilde:
			argv = append(argv, p.parseWord())
			continue
		default:
			break
		}
		return &ast.Cmd{Argv: argv}
	}
}

func (p *Parser) parseWord() *ast.Word {
	var exprs []ast.Expr
	var str strings.Builder
	for {
		switch l := p.peek(); l.tok {
		case token.Newline:
			if str.Len() > 0 {
				// We're inside a multi-line string, and expect
				// more of the string on the next line.
				p.done <- false
			} else {
				return &ast.Word{SubExprs: exprs}
			}
		case token.EscapedNewline:
			p.done <- false
		case token.String:
			str.WriteString(l.text)
			exprs = append(exprs, ast.String{Text: str.String()})
			str.Reset()
		case token.SubString:
			str.WriteString(l.text)
		case token.Tilde:
			exprs = append(exprs, ast.Tilde{Text: l.text})
		default:
			if str.Len() > 0 {
				panic(fmt.Sprintf(
					"parser: unexpected token %v", l))
			} else {
				return &ast.Word{SubExprs: exprs}
			}
		}
		p.next()
	}
}
