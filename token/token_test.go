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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenString(t *testing.T) {
	set := make(map[string]struct{})
	for tok := Token(tokenBegin + 1); tok < tokenEnd; tok++ {
		// Check that every valid token can be converted to a string.
		// This catches cases where a new token const is added but the
		// associated case statement in Token.String() is forgotten.
		assert.NotPanics(t, func() {
			str := tok.String()
			assert.NotEmpty(t, str)
			// Check that every token's string representation is
			// unqiue. Hopefully this catches any copy/paste errors.
			assert.NotContains(t, set, str)
			set[str] = struct{}{}
		})
	}
}

func TestTokenStringPanicsIfInvalid(t *testing.T) {
	var tok Token = -1
	assert.Panics(t, func() { _ = tok.String() })
}
