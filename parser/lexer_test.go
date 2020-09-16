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

	"github.com/meshshell/mesh/token"
)

func TestLexerStrings(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output []string
	}{
		{"Command", "ls -l", []string{"ls", "-l"}},
		{"ExtraSpaces", ` a  b\ c  `, []string{"a", "b c"}},
		{"SingleQuoted", `a 'b  c\'"'`, []string{"a", `b  c'"`}},
		{"DoubleQuoted", `a "b  c'\""`, []string{"a", `b  c'"`}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, ch := lex(test.name, test.input)
			for _, expected := range test.output {
				actual := <-ch
				assert.Equal(t, token.String, actual.tok)
				assert.Equal(t, expected, actual.text)
			}
		})
	}
}
