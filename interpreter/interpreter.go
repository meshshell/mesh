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
	"os/exec"

	"github.com/meshshell/mesh/ast"
)

type Interpreter struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func (i *Interpreter) VisitCmd(c *ast.Cmd) (int, error) {
	if b, ok := newBuiltin(c.Name, c.Args); ok {
		if err := b.run(); err != nil {
			return 1, err
		}
		return 0, nil
	}
	cmd := exec.Command(c.Name, c.Args...)
	cmd.Stdin = i.Stdin
	cmd.Stdout = i.Stdout
	cmd.Stderr = i.Stderr
	err := cmd.Run()
	status := cmd.ProcessState.ExitCode()
	return status, err
}
