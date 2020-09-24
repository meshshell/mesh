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
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meshshell/mesh/ast"
)

func TestInterpreter(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name    string
		argv    []string
		success bool
	}{
		{"NopCommand", []string{}, true},
		{"BuiltinSucceeds", []string{"cd", home}, true},
		{"BuiltinFails", []string{"cd", os.DevNull}, false},
		{"NormalCommand", []string{"true"}, true},
	}

	stdin, err := os.Open(os.DevNull)
	require.NoError(t, err)
	defer stdin.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr strings.Builder
			interp := Interpreter{stdin, &stdout, &stderr}
			var exprs []ast.Expr
			for _, text := range test.argv {
				exprs = append(exprs, ast.String{Text: text})
			}
			cmd := &ast.Cmd{Argv: exprs}
			status, err := interp.VisitCmd(cmd)
			if test.success {
				assert.Equal(t, 0, status)
				assert.NoError(t, err)
			} else {
				assert.NotEqual(t, 0, status)
				assert.Error(t, err)
			}
		})
	}
}
