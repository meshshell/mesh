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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meshshell/mesh/ast"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name   string
		inputs []string
		cmd    string
		args   []string
	}{
		{"EmptyString", []string{""}, "", []string{}},
		{"OneWord", []string{"ls"}, "ls", []string{}},
		{"TwoWords", []string{"ls -l"}, "ls", []string{"-l"}},
		{"TwoLines", []string{"a 'b", "c'"}, "a", []string{"b\nc"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := NewParser(t.Name())
			for i, line := range test.inputs {
				isLastLine := i == len(test.inputs)-1
				assert.Equal(t, isLastLine, p.Parse(line))
			}
			stmt, err := p.Result()
			require.NoError(t, err)
			cmd, ok := stmt.(*ast.Cmd)
			require.True(t, ok)
			assert.Equal(t, test.cmd, cmd.Name)
			assert.Equal(t, test.args, cmd.Args)
		})
	}
}

func TestParserResultWhileLocked(t *testing.T) {
	p := NewParser(t.Name())
	require.False(t, p.Parse("echo 'asdf"))
	assert.Panics(t, func() { p.Result() })
}
