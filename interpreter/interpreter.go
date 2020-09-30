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
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

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

func (shell *Interpreter) VisitPipeline(p *ast.Pipeline) (int, error) {
	var fromPipe io.ReadCloser
	statuses := make([]int, len(p.Stmts))
	errs := make([]error, len(p.Stmts))
	var wg sync.WaitGroup
	wg.Add(len(p.Stmts))
	for index, stmt := range p.Stmts {
		subshell := &Interpreter{Stderr: shell.Stderr}
		if index == 0 {
			// First command in the pipeline, so read from stdin.
			subshell.Stdin = shell.Stdin
		} else {
			// Otherwise read from a pipe. The output side of the
			// pipe will have been created in the previous iteration
			// of this loop.
			subshell.Stdin = fromPipe
		}
		var toPipe io.WriteCloser
		if index == len(p.Stmts)-1 {
			// Last command in the pipeline, so write to stdout.
			subshell.Stdout = shell.Stdout
		} else {
			// Otherwise create a pipe and write to it.
			var pipeErr error
			fromPipe, toPipe, pipeErr = os.Pipe()
			if pipeErr != nil {
				return -1, pipeErr
			}
			defer fromPipe.Close()
			subshell.Stdout = toPipe
		}
		go func(index int, stmt ast.Stmt) {
			// VisitCmd runs synchronously, so run it in a goroutine
			// to ensure that the pipeline runs concurrently.
			statuses[index], errs[index] = stmt.Visit(subshell)
			if toPipe != nil {
				// Close the write-side of the pipe, so that the
				// next command in the pipeline doesn't block
				// trying to read from the pipe.
				toPipe.Close()
			}
			wg.Done()
		}(index, stmt)
	}
	wg.Wait()
	// TODO: implement `pipefail` behaviour?
	return statuses[len(p.Stmts)-1], errs[len(p.Stmts)-1]
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
