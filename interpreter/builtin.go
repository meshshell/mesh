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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type builtin struct {
	fn   func(*builtin) error
	args []string
}

func newBuiltin(name string, args []string) (*builtin, bool) {
	switch name {
	case "cd":
		return &builtin{fn: cd, args: args}, true
	case "exit":
		return &builtin{fn: exit, args: args}, true
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
		if target == "-" {
			var ok bool
			target, ok = os.LookupEnv("OLDPWD")
			if !ok {
				return fmt.Errorf("cd: OLDPWD not set")
			}
		}
	default:
		return errors.New("cd: too many arguments")
	}
	oldpwd := os.Getenv("PWD")
	newpwd, _ := filepath.Abs(target)
	if err := os.Chdir(target); err != nil {
		return fmt.Errorf("cd: %w", err)
	}
	os.Setenv("OLDPWD", oldpwd)
	return os.Setenv("PWD", newpwd)
}

type ExitStatus int

func (e ExitStatus) Error() string {
	return fmt.Sprintf("exit %d", int(e))
}

func exit(b *builtin) error {
	switch len(b.args) {
	case 0:
		return ExitStatus(0)
	case 1:
		i, err := strconv.Atoi(b.args[0])
		if err != nil {
			return errors.New("exit: integer argument required")
		}
		return ExitStatus(i)
	default:
		return errors.New("exit: too many arguments")
	}
}
