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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type integrationTest struct {
	name   string
	script string
	status int
	stdout string
	stderr string
}

func (test *integrationTest) run(t *testing.T) {
	stdin := mustOpen(t, os.DevNull)
	var stdout, stderr strings.Builder
	s := newNonInteractive(strings.NewReader(test.script))
	status := repl(test.name, s, &stdio{stdin, &stdout, &stderr})
	assert.Equal(t, test.status, status)
	assert.Equal(t, test.stdout, stdout.String())
	assert.Equal(t, test.stderr, stderr.String())
}

func TestTildeExpansion(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	for _, test := range []integrationTest{
		{
			name:   "Tilde",
			script: "echo ~\n",
			stdout: home + "\n",
		}, {
			name:   "TildeWithSubDirs",
			script: "echo " + filepath.Join("~", "Desktop") + "\n",
			stdout: filepath.Join(home, "Desktop") + "\n",
		}, {
			name:   "TildeInsideQuotes",
			script: "echo '~'\n",
			stdout: "~\n",
		}, {
			name:   "TildeInsideString",
			script: "echo x~\n",
			stdout: "x~\n",
		},
	} {
		t.Run(test.name, test.run)
	}
}
