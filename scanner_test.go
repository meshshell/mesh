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

package main

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNonInteractive(t *testing.T) {
	n := newNonInteractive(strings.NewReader("one\ntwo"))

	// These functions are no-ops, but run them anyway so that we get points
	// for test coverage.
	n.setIgnoreEOF(false)
	n.setPrompt("")
	n.setViMode(false)

	line, err := n.readLine()
	assert.NoError(t, err)
	assert.Equal(t, "one", line)
	line, err = n.readLine()
	assert.NoError(t, err)
	assert.Equal(t, "two", line)
	_, err = n.readLine()
	assert.Equal(t, io.EOF, err)
}
