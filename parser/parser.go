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

func newParserError(format string, a ...interface{}) parserError {
	return parserError{fmt.Sprintf(format, a...)}
}

func (pe parserError) Error() string {
	return "parser: " + pe.msg
}

type Parser struct {
	lex    *lexer
	done   chan bool
	lock   sync.Mutex
	locked bool
	stmt   ast.Stmt
	err    error
	curr   *lexeme
}

func NewParser(filename string) *Parser {
	return &Parser{
		lex:  newLexer(filename),
		done: make(chan bool),
	}
}

func (p *Parser) Parse(line string) bool {
	if !p.locked {
		go p.parseStmtList()
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

// accept consumes the current token, so that the accept call to peek() or
// trim() will return a new token
func (p *Parser) accept() {
	if p.curr == nil {
		// TODO: If this panic happens, it's a bug, and we should prompt
		// the user to report it (and probably provide more info about
		// what went wrong, such as the next token). This function must
		// only ever be called after a call to peek() or trim().
		panic("parser: tried to skip over unseen token")
	}
	p.curr = nil
}

// peek returns the current token, retrieving it from the lexer if necessary
func (p *Parser) peek() *lexeme {
	if p.curr == nil {
		l := <-p.lex.lexemes
		p.curr = &l
	}
	return p.curr
}

// trim is like peek(), except that it consumes any whitespace before returning
// the current token
func (p *Parser) trim() *lexeme {
	for {
		switch p.peek().tok {
		case token.EscapedNewline:
			p.done <- false
			fallthrough
		case token.Whitespace:
			p.accept()
			continue
		default:
			return p.curr
		}
	}
}

func (p *Parser) parseStmtList() {
	p.lock.Lock()
	p.locked = true
	p.stmt, p.err, p.curr = nil, nil, nil
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(parserError)
			if !ok {
				panic(r)
			}
			p.err = err
			// If the parser panics before parsing the current line,
			// the lexer will still continue to run. So we need to
			// drain the p.lexemes channel of all tokens until the
			// end of the line, so that the lexer doesn't block.
			for p.peek().tok != token.Newline &&
				p.peek().tok != token.EscapedNewline {
				p.accept()
			}
			p.accept()
		}
		p.locked = false
		p.done <- true
		p.lock.Unlock()
	}()
	var stmts []ast.Stmt
	for {
		switch l := p.trim(); l.tok {
		case token.Newline:
			p.accept()
			p.stmt = &ast.StmtList{Stmts: stmts}
			return
		case token.Semicolon:
			p.accept()
			continue
		default:
			stmts = append(stmts, p.parseStmt())
		}
	}
}

func (p *Parser) parseStmt() ast.Stmt {
	switch l := p.trim(); l.tok {
	case token.Dollar:
		panic(newParserError("assignment stmt not yet implemented"))
	case token.String, token.SubString, token.Tilde:
		return p.parsePipeline()
	case token.Semicolon, token.Newline:
		return &ast.Cmd{Argv: []ast.Expr{}}
	default:
		panic(newParserError("unexpected token: %v", l))
	}
}

func (p *Parser) parsePipeline() *ast.Pipeline {
	stmts := []ast.Stmt{p.parseCmd()}
	for {
		switch l := p.trim(); l.tok {
		case token.Pipe:
			p.accept()
		case token.Semicolon, token.Newline:
			return &ast.Pipeline{Stmts: stmts}
		default:
			stmts = append(stmts, p.parseCmd())
		}
	}
}

func (p *Parser) parseCmd() *ast.Cmd {
	var argv []ast.Expr
	for {
		switch l := p.trim(); l.tok {
		case token.String, token.SubString, token.Dollar, token.Tilde:
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
				p.accept()
			} else {
				return &ast.Word{SubExprs: exprs}
			}
		case token.String:
			str.WriteString(l.text)
			exprs = append(exprs, ast.String{Text: str.String()})
			str.Reset()
			p.accept()
		case token.SubString:
			str.WriteString(l.text)
			p.accept()
		case token.Dollar:
			p.accept()
			v := p.parseVar()
			if v == nil {
				// The `$` was not followed by a valid
				// identifier, so just treat it as literal text.
				exprs = append(exprs, ast.String{Text: l.text})
			} else {
				exprs = append(exprs, v)
			}
		case token.Tilde:
			exprs = append(exprs, ast.Tilde{Text: l.text})
			p.accept()
		default:
			if str.Len() > 0 {
				panic(newParserError(
					"parser: unexpected token %v", l))
			} else {
				return &ast.Word{SubExprs: exprs}
			}
		}
	}
}

func (p *Parser) parseVar() *ast.Var {
	// TODO: Allow arrays to be indexed, and maps to be looked up.
	switch l := p.peek(); l.tok {
	case token.Identifier:
		p.accept()
		return &ast.Var{Identifier: l.text}
	default:
		return nil
	}
}
