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
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createFile(t *testing.T, contents string) string {
	assert.Equal(t, "\n", contents[len(contents)-1:], "no trailing newline")
	f, err := ioutil.TempFile("", "mesh")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.Remove(f.Name())) })
	defer f.Close()
	n, err := f.WriteString(contents)
	require.NoError(t, err)
	require.Equal(t, len(contents), n)
	return f.Name()
}

func mustOpen(t *testing.T, name string) *os.File {
	f, err := os.Open(name)
	require.NoError(t, err)
	return f
}

func TestCommandFromArgs(t *testing.T) {
	stdin := mustOpen(t, os.DevNull)
	var stdout, stderr strings.Builder
	status := mesh(
		"mesh",
		[]string{"-c", "echo foo"},
		&stdio{stdin, &stdout, &stderr},
	)
	assert.Equal(t, 0, status)
	assert.Equal(t, "foo\n", stdout.String())
	assert.Empty(t, stderr.String())
}

func TestScriptFromFile(t *testing.T) {
	stdin := mustOpen(t, os.DevNull)
	var stdout, stderr strings.Builder
	status := mesh(
		"mesh",
		[]string{createFile(t, "echo bar\n")},
		&stdio{stdin, &stdout, &stderr},
	)
	assert.Equal(t, 0, status)
	assert.Equal(t, "bar\n", stdout.String())
	assert.Empty(t, stderr.String())
}

func TestScriptFromStdin(t *testing.T) {
	stdin := mustOpen(t, createFile(t, "echo baz\n"))
	var stdout, stderr strings.Builder
	status := mesh("mesh", []string{}, &stdio{stdin, &stdout, &stderr})
	assert.Equal(t, 0, status)
	assert.Equal(t, "baz\n", stdout.String())
	assert.Empty(t, stderr.String())
}

func TestErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		arg           string
		shouldSucceed bool
	}{
		{"ShortHelpFlag", "-h", true},
		{"LongHelpFlag", "-help", true},
		{"BadFlag", "-badflag", false},
		{"NonExistentScript", "/nonexistent", false},
		{"ParseError", "-c=|", false},
		{"ExecError", "-c=/nonexistent", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stdin := mustOpen(t, os.DevNull)
			var stdout, stderr strings.Builder
			status := mesh(
				"mesh",
				[]string{test.arg},
				&stdio{stdin, &stdout, &stderr},
			)
			if test.shouldSucceed {
				assert.Equal(t, 0, status)
			} else {
				assert.NotEqual(t, 0, status)
			}
			assert.NotEmpty(t, stderr.String())
		})
	}
}
