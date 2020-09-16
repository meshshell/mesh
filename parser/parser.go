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
	"errors"
	"fmt"

	"github.com/meshshell/mesh/ast"
	"github.com/meshshell/mesh/token"
)

type Parser struct {
	filename string
}

func NewParser(filename string) *Parser {
	return &Parser{filename}
}

func (p *Parser) Parse(line string) (ast.Stmt, error) {
	_, lexemes := lex(p.filename, line)
	var argv []string
	for lexeme := range lexemes {
		if lexeme.tok == token.SubString1 ||
			lexeme.tok == token.SubString2 {
			msg := "multi-line strings are not yet implemented"
			return nil, fmt.Errorf("parser: %s", msg)
		} else if lexeme.tok != token.String {
			return nil, fmt.Errorf(
				"parser: unexpected token %q", lexeme.text)
		} else if lexeme.text == "|" {
			return nil, errors.New(
				"parser: pipes are not yet implemented")
		}
		argv = append(argv, lexeme.text)
	}
	if len(argv) == 0 {
		return &ast.Cmd{}, nil
	}
	return &ast.Cmd{Name: argv[0], Args: argv[1:]}, nil
}
