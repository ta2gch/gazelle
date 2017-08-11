// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package runtime

import (
	"github.com/ta2gch/iris/runtime/environment"
	"github.com/ta2gch/iris/runtime/ilos"
	"github.com/ta2gch/iris/runtime/ilos/class"
	"github.com/ta2gch/iris/runtime/ilos/instance"
)

// Functionp returns t if obj is a (normal or generic) function;
// otherwise, returns nil. obj may be any ISLISP object.
//
// Function bindings are entities established during execution of
// a prepared labels or flet forms or by a function-defining form.
// A function binding is an association between an identifier, function-name,
// and a function object that is denoted by function-name—if in operator
// position—or by (function function-name) elsewhere.
func Functionp(local, global *environment.Environment, fun ilos.Instance) (ilos.Instance, ilos.Instance) {
	if instance.Of(class.Function, fun) {
		return T, nil
	}
	return Nil, nil
}

// Function returns the function object named by function-name.
//
// An error shall be signaled if no binding has been established for the identifier
// in the function namespace of current lexical environment (error-id. undefined-function).
// The consequences are undefined if the function-name names a macro or special form
func Function(local, global *environment.Environment, fun ilos.Instance) (ilos.Instance, ilos.Instance) {
	// car must be a symbol
	if !instance.Of(class.Symbol, fun) {
		return nil, instance.New(class.DomainError, map[string]ilos.Instance{
			"OBJECT":         fun,
			"EXPECTED-CLASS": class.Symbol,
		})
	}
	if f, ok := local.Function.Get(fun); ok {
		return f, nil
	}
	if f, ok := global.Function.Get(fun); ok {
		return f, nil
	}
	return nil, instance.New(class.UndefinedFunction, map[string]ilos.Instance{
		"NAME":      fun,
		"NAMESPACE": instance.New(class.Symbol, "FUNCTION"),
	})
}

func checkLambdaList(lambdaList ilos.Instance) ilos.Instance {
	cdr := lambdaList
	ok := false
	for instance.Of(class.Cons, cdr) {
		cadr := instance.UnsafeCar(cdr)
		cddr := instance.UnsafeCdr(cdr)
		if !instance.Of(class.Symbol, cadr) {
			break
		}
		if cadr == instance.New(class.Symbol, ":REST") || cadr == instance.New(class.Symbol, "&REST") {
			if instance.Of(class.Cons, cddr) && instance.Of(class.Symbol, instance.UnsafeCar(cddr)) && instance.Of(class.Null, instance.UnsafeCdr(cddr)) {
				ok = true
			}
			break
		}
		cdr = cddr
	}
	if !ok && cdr == Nil {
		ok = true
	}
	if !ok {
		return instance.New(class.ProgramError)
	}
	return nil
}

// Lambda special form creates a function object.
//
// The scope of the identifiers of the lambda-list is the sequence of forms form*,
// collectively referred to as the body.
//
// When the prepared function is activated later (even if transported as object
// to some other activation) with some arguments, the body of the function is
// evaluated as if it was at the same textual position where the lambda special
// form is located, but in a context where the lambda variables are bound
// in the variable namespace with the values of the corresponding arguments.
// A &rest or :rest variable, if any, is bound to the list of the values of
// the remaining arguments. An error shall be signaled if the number of
// arguments received is incompatible with the specified lambda-list
// (error-id. arity-error).
//
// Once the lambda variables have been bound, the body is executed.
// If the body is empty, nil is returned otherwise the result of the evaluation of
// the last form of body is returned if the body was not left by a non-local exit.
//
// If the function receives a &rest or :rest parameter R, the list L1 to which that
// parameter is bound has indefinite extent. L1 is newly allocated unless the function
// was called with apply and R corresponds to the final argument, L2 , to that call
// to apply (or some subtail of L2), in which case it is implementation defined whether
// L1 shares structure with L2 .
func Lambda(local, global *environment.Environment, lambdaList ilos.Instance, form ...ilos.Instance) (ilos.Instance, ilos.Instance) {
	if err := checkLambdaList(lambdaList); err != nil {
		return nil, err
	}
	return newNamedFunction(local, global, instance.New(class.Symbol, "ANONYMOUS-FUNCTION"), lambdaList, form...), nil
}

// Labels special form allow the definition of new identifiers in the function
// namespace for function objects.
//
// In a labels special form the scope of an identifier function-name is the whole
// labels special form (excluding nested scopes, if any); for the flet special form,
// the scope of an identifier is only the body-form*. Within these scopes,
// each function-name is bound to a function object whose behavior is equivalent
// to (lambda lambda-list form*), where free identifier references are resolved as follows:
//
// For a labels form, such free references are resolved in the lexical environment
// that was active immediately outside the labels form augmented by the function
// bindings for the given function-names (i.e., any reference to a function
// function-name refers to a binding created by the labels).
//
// For a flet form, free identifier references in the lambda-expression are resolved
// in the lexical environment that was active immediately outside the flet form
// (i.e., any reference to a function function-name are not visible).
//
// During activation, the prepared labels or flet establishes function bindings and
// then evaluates each body-form in the body sequentially; the value of the last one
// (or nil if there is none) is the value returned by the special form activation.
//
// No function-name may appear more than once in the function bindings.
func Labels(local, global *environment.Environment, functions ilos.Instance, bodyForm ...ilos.Instance) (ilos.Instance, ilos.Instance) {
	cdr := functions
	for instance.Of(class.Cons, cdr) {
		cadr := instance.UnsafeCar(cdr)
		functionName := instance.UnsafeCar(cadr)
		lambdaList := instance.UnsafeCar(instance.UnsafeCdr(cadr))
		if err := checkLambdaList(lambdaList); err != nil {
			return nil, err
		}

		cddadr := instance.UnsafeCdr(instance.UnsafeCdr(cadr))
		form := []ilos.Instance{}
		for instance.Of(class.Cons, cddadr) {
			caddadr := instance.UnsafeCar(cddadr)
			form = append(form, caddadr)
			cddadr = instance.UnsafeCdr(cddadr)
		}
		local.Function.Define(functionName, newNamedFunction(local, global, functionName, lambdaList, form...))
		cdr = instance.UnsafeCdr(cdr)
	}
	ret := Nil
	var err ilos.Instance
	for _, form := range bodyForm {
		ret, err = Eval(local, global, form)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

// Flet special form allow the definition of new identifiers in the function
// namespace for function objects (see Labels).
func Flet(local, global *environment.Environment, functions ilos.Instance, bodyForm ...ilos.Instance) (ilos.Instance, ilos.Instance) {
	cdr := functions
	env := environment.New()
	env.BlockTag = append(local.BlockTag, env.BlockTag...)
	env.TagbodyTag = append(local.TagbodyTag, env.TagbodyTag...)
	env.CatchTag = append(local.CatchTag, env.CatchTag...)
	env.Variable = append(local.Variable, env.Variable...)
	env.Function = append(local.Function, env.Function...)
	env.Special = append(local.Special, env.Special...)
	env.Macro = append(local.Macro, env.Macro...)
	env.DynamicVariable = append(local.DynamicVariable, env.DynamicVariable...)
	for instance.Of(class.Cons, cdr) {
		cadr := instance.UnsafeCar(cdr)
		functionName := instance.UnsafeCar(cadr)
		lambdaList := instance.UnsafeCar(instance.UnsafeCdr(cadr))
		if err := checkLambdaList(lambdaList); err != nil {
			return nil, err
		}

		cddadr := instance.UnsafeCdr(instance.UnsafeCdr(cadr))
		form := []ilos.Instance{}
		for instance.Of(class.Cons, cddadr) {
			caddadr := instance.UnsafeCar(cddadr)
			form = append(form, caddadr)
			cddadr = instance.UnsafeCdr(cddadr)
		}
		env.Function.Define(functionName, newNamedFunction(local, global, functionName, lambdaList, form...))
		cdr = instance.UnsafeCdr(cdr)
	}
	ret := Nil
	var err ilos.Instance
	for _, form := range bodyForm {
		ret, err = Eval(env, global, form)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

// Apply applies function to the arguments, obj*, followed by the elements of list,
// if any. It returns the value returned by function.
//
// An error shall be signaled if function is not a function (error-id. domain-error).
// Each obj may be any ISLISP object. An error shall be signaled
// if list is not a proper list (error-id. improper-argument-list).
func Apply(local, global *environment.Environment, function ilos.Instance, obj ...ilos.Instance) (ilos.Instance, ilos.Instance) {
	list := Nil
	if instance.Of(class.List, obj[len(obj)-1]) {
		list = obj[len(obj)-1]
		if !isProperList(list) {
			return nil, instance.New(class.ProgramError)
		}
		obj = obj[:len(obj)-1]
	}
	for i := len(obj) - 1; i >= 0; i-- {
		list = instance.New(class.Cons, obj[i], list)
	}
	if !instance.Of(class.Function, function) {
		return nil, instance.New(class.DomainError, map[string]ilos.Instance{
			"OBJECT":         function,
			"EXPECTED-CLASS": class.Function,
		})
	}
	ret, err := function.(instance.Applicable).Apply(local, global, list)
	return ret, err
}

// Funcall activates the specified function function and returns the value that the function returns.
// The ith argument (2 ≤ i) of funcall becomes the (i − 1)th argument of the function.
//
// An error shall be signaled if function is not a function (error-id. domain-error).
// Each obj may be any ISLISP object.
func Funcall(local, global *environment.Environment, function ilos.Instance, obj ...ilos.Instance) (ilos.Instance, ilos.Instance) {
	ret, err := Apply(local, global, function, obj...)
	return ret, err
}
