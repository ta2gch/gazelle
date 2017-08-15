// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package instance

import (
	"github.com/ta2gch/iris/runtime/ilos"
	"github.com/ta2gch/iris/runtime/ilos/class"
)

func NewArithmeticError(operation, operands ilos.Instance) ilos.Instance {
	return New(class.ArithmeticError, map[string]ilos.Instance{
		"OPERATION": operation,
		"OPERANDS":  operands,
	})
}

func NewParseError(str, expectedClass ilos.Instance) ilos.Instance {
	return New(class.ParseError, map[string]ilos.Instance{
		"STRING":          str,
		"EXPECTED-CLASSS": expectedClass,
	})
}

func NewDomainError(object, expectedClass ilos.Instance) ilos.Instance {
	return New(class.DomainError, map[string]ilos.Instance{"CAUSE": Symbol("DOMAIN-ERROR"),
		"OBJECT":          object,
		"EXPECTED-CLASSS": expectedClass,
	})
}

func NewUndefinedFunction(name ilos.Instance) ilos.Instance {
	return New(class.UndefinedFunction, map[string]ilos.Instance{
		"NAME":      name,
		"NAMESPACE": Symbol("FUNCTION"),
	})
}

func NewUndefinedVariable(name ilos.Instance) ilos.Instance {
	return New(class.UndefinedVariable, map[string]ilos.Instance{
		"NAME":      name,
		"NAMESPACE": Symbol("VARIABLE"),
	})
}

func NewArityError() ilos.Instance {
	return New(class.ProgramError, map[string]ilos.Instance{})
}

func NewIndexOutOfRange() ilos.Instance {
	return New(class.ProgramError, map[string]ilos.Instance{})
}

func NewImmutableBinding() ilos.Instance {
	return New(class.ProgramError, map[string]ilos.Instance{})
}