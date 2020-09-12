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
	"fmt"
	"errors"
	"os"
)

type builtin struct {
	fn   func(*builtin) error
	args []string
}

func newBuiltin(name string, args []string) (*builtin, bool) {
	switch name {
	case "cd":
		return &builtin{fn: cd, args: args}, true
	default:
		return nil, false
	}
}

func (b *builtin) run() error {
	return b.fn(b)
}

func cd(b *builtin) error {
	var target string
	switch len(b.args) {
	case 0:
		var err error
		target, err = os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cd: %w", err)
		}
	case 1:
		target = b.args[0]
	default:
		return errors.New("cd: too many arguments")
	}
	if err := os.Chdir(target); err != nil {
		return fmt.Errorf("cd: %w", err)
	}
	return nil
}
