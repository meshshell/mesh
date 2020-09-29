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

package interpreter

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/meshshell/mesh/ast"
)

type Interpreter struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func (i *Interpreter) VisitStmtList(s *ast.StmtList) (int, error) {
	var status int
	var err error
	for _, stmt := range s.Stmts {
		if status, err = stmt.Visit(i); err != nil {
			return status, err
		}
	}
	return status, err
}

func (i *Interpreter) VisitPipeline(p *ast.Pipeline) (int, error) {
	return 0, errors.New("not yet implemented")
}

func (i *Interpreter) VisitCmd(c *ast.Cmd) (int, error) {
	var argv []string
	for _, expr := range c.Argv {
		text, err := expr.Visit(i)
		if err != nil {
			return -1, err
		}
		argv = append(argv, text)
	}
	if len(argv) == 0 {
		return 0, nil
	} else if b, ok := newBuiltin(argv[0], argv[1:]); ok {
		if err := b.run(); err != nil {
			return 1, err
		}
		return 0, nil
	} else {
		cmd := exec.Command(argv[0], argv[1:]...)
		cmd.Stdin = i.Stdin
		cmd.Stdout = i.Stdout
		cmd.Stderr = i.Stderr
		err := cmd.Run()
		status := cmd.ProcessState.ExitCode()
		return status, err
	}
}

func (i *Interpreter) VisitString(s ast.String) (string, error) {
	return s.Text, nil
}

func (i *Interpreter) VisitTilde(t ast.Tilde) (string, error) {
	home, err := os.UserHomeDir()
	return home, err
}

func (i *Interpreter) VisitVar(v ast.Var) (string, error) {
	// TODO: Implement an internal symbol table for shell variables.
	return os.Getenv(v.Identifier), nil
}

func (i *Interpreter) VisitWord(w ast.Word) (string, error) {
	var word strings.Builder
	for _, subExpr := range w.SubExprs {
		s, err := subExpr.Visit(i)
		if err != nil {
			return "", err
		}
		word.WriteString(s)
	}
	return word.String(), nil
}
