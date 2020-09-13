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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltinCD(t *testing.T) {
	homedir, err := os.UserHomeDir()
	require.NoError(t, err)
	tempdir, err := ioutil.TempDir("", "mesh")
	require.NoError(t, err)
	defer os.RemoveAll(tempdir)
	nonexistent := filepath.Join(tempdir, "nonexistent")

	tests := []struct {
		name   string
		args   []string
		target string
	}{
		{"NoArgs", []string{}, homedir},
		{"ToTempDir", []string{tempdir}, tempdir},
		{"NonExistent", []string{nonexistent}, ""},
		{"TooManyArgs", []string{"too", "many"}, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b, ok := newBuiltin("cd", test.args)
			require.True(t, ok)
			err := b.run()
			if test.target == "" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				dir, err := os.Getwd()
				require.NoError(t, err)
				assert.Equal(t, test.target, dir)
			}
		})
	}
}

func TestExitStatusError(t *testing.T) {
	assert.Equal(t, "exit 2", ExitStatus(2).Error())
}
