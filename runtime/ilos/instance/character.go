// This Source Code Form is subject to the terms of the Mozilla Public License,
// v. 2.0. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/.

package instance

import (
	"github.com/islisp-dev/iris/runtime/ilos"
)

// Character

type Character rune

func NewCharacter(r rune) ilos.Instance {
	return Character(r)
}

func (Character) Class() ilos.Class {
	return CharacterClass
}

func (i Character) String() string {
	switch rune(i) {
	case ' ':
		return `#\SPACE`
	case '\n':
		return `#\NEWLINE`
	default:
		return `#\` + string(i)
	}
}
