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

type Parser struct {
	lex    *lexer
	done   chan bool
	lock   sync.Mutex
	locked bool
	stmt   ast.Stmt
	err    error
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

func (p *Parser) parseStmt() {
	p.lock.Lock()
	p.locked = true
	p.stmt, p.err = nil, nil
	defer func() {
		p.locked = false
		p.done <- true
		p.lock.Unlock()
	}()

	var argv []string
	var tmp strings.Builder
	escaped := false
	for lexeme := range p.lex.lexemes {
		switch lexeme.tok {
		case token.SubString:
			tmp.WriteString(lexeme.text)
		case token.String:
			if tmp.Len() == 0 {
				argv = append(argv, lexeme.text)
			} else {
				tmp.WriteString(lexeme.text)
				argv = append(argv, tmp.String())
				tmp.Reset()
			}
		case token.Escape:
			escaped = true
		case token.Newline:
			if escaped {
				escaped = false
				p.done <- false
			} else if tmp.Len() == 0 {
				if len(argv) == 0 {
					argv = []string{""}
				}
				p.stmt = &ast.Cmd{Name: argv[0], Args: argv[1:]}
				return
			} else {
				p.done <- false
			}
		default:
			p.err = fmt.Errorf(
				"parser: unexpected token %q", lexeme.text)
			continue
		}
	}
}
