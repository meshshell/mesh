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

package token

import (
	"fmt"
)

type Token int

const (
	tokenBegin = iota

	Newline
	EscapedNewline
	Whitespace

	Identifier
	String
	SubString

	Dollar
	Pipe
	Semicolon
	Tilde

	tokenEnd
)

func (t Token) String() string {
	switch t {
	case Newline:
		return "Newline"
	case EscapedNewline:
		return "EscapedNewline"
	case Whitespace:
		return "Whitespace"
	case Identifier:
		return "Identifier"
	case String:
		return "String"
	case SubString:
		return "SubString"
	case Dollar:
		return "Dollar"
	case Pipe:
		return "Pipe"
	case Semicolon:
		return "Semicolon"
	case Tilde:
		return "Tilde"
	default:
		panic(fmt.Sprintf("invalid token.Token: %d", t))
	}
}
