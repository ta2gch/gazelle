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

/*
ISLISP defines three ways in which to perform non-local exits:
  Destination   Kind     Established by  Invoked by  Operation Performed
  block name    block    return-from     lexical     exit
  tagbody tag   tagbody  go              lexical     transfer of control
  catch tag     atch     throw           dynamic     exit

A non-local exit, is an operation that forces transfer of control and possibly
data from an invoking special form to a previously established point in a program,
called the destination of the exit.

A lexical exit is a non-local exit from a return-from form to a block form
which contains it both lexically and dynamically, forcing the block to return
an object specified in the return-from form.

A dynamic exit is a non-local exit from a throw form to a catch form
which contains it dynamically (but not necessarily lexically),
forcing the catch to return an object specified in the throw form.

A lexical transfer of control is a non-local exit from a go form to a tagged point
in a tagbody form which contains it both lexically and dynamically.

When a non-local exit is initiated, any potential destination that was established
more recently than the destination to which control is being transferred
is immediately considered invalid.
*/

func Block(local environment.Environment, tag ilos.Instance, body ...ilos.Instance) (ilos.Instance, ilos.Instance) {
	var err ilos.Instance
	tag, err = Eval(local, tag) // Checked at the top of// This function
	if err != nil {
		return nil, err
	}
	if ilos.InstanceOf(class.Number, tag) || ilos.InstanceOf(class.Character, tag) {
		return nil, instance.NewDomainError(tag, class.Object)
	}
	if !local.BlockTag.Define(tag, nil) {
		return nil, instance.NewImmutableBinding()
	}
	var fail ilos.Instance
	sucess := Nil
	for _, cadr := range body {
		sucess, fail = Eval(local, cadr)
		if fail != nil {
			if ilos.InstanceOf(class.BlockTag, fail) {
				tag1, _ := fail.(instance.Instance).GetSlotValue(instance.NewSymbol("TAG"), class.Escape) // Checked at the head of// This condition
				if tag == tag1 {
					obj, _ := fail.(instance.Instance).GetSlotValue(instance.NewSymbol("OBJECT"), class.BlockTag) // Checked at the head of// This condition
					return obj, nil
				}
			}
			return nil, fail
		}
	}
	return sucess, nil
}

func ReturnFrom(local environment.Environment, tag, object ilos.Instance) (ilos.Instance, ilos.Instance) {
	var err ilos.Instance
	tag, err = Eval(local, tag)
	if err != nil {
		return nil, err
	}
	if ilos.InstanceOf(class.Number, tag) || ilos.InstanceOf(class.Character, tag) {
		return nil, instance.NewDomainError(tag, class.Object)
	}
	object, err = Eval(local, object)
	if err != nil {
		return nil, err
	}
	if _, ok := local.BlockTag.Get(tag); !ok {
		return nil, instance.NewSimpleError(instance.NewString("%v is not defined as the tag"), tag)
	}
	return nil, instance.NewBlockTag(tag, object)
}

func Catch(local environment.Environment, tag ilos.Instance, body ...ilos.Instance) (ilos.Instance, ilos.Instance) {
	var err ilos.Instance
	tag, err = Eval(local, tag)
	if err != nil {
		return nil, err
	}
	if ilos.InstanceOf(class.Number, tag) || ilos.InstanceOf(class.Character, tag) {
		return nil, instance.NewDomainError(tag, class.Object)
	}
	if !local.CatchTag.Define(tag, nil) {
		return nil, instance.NewImmutableBinding()
	}
	var fail ilos.Instance
	sucess := Nil
	for _, cadr := range body {
		sucess, fail = Eval(local, cadr)
		if fail != nil {
			if ilos.InstanceOf(class.CatchTag, fail) {
				tag1, _ := fail.(instance.Instance).GetSlotValue(instance.NewSymbol("TAG"), class.Escape) // Checked at the head of// This condition
				if tag == tag1 {
					obj, _ := fail.(instance.Instance).GetSlotValue(instance.NewSymbol("OBJECT"), class.CatchTag) // Checked at the head of// This condition
					return obj, nil
				}
			}
			return nil, fail
		}
	}
	return sucess, nil
}

func Throw(local environment.Environment, tag, object ilos.Instance) (ilos.Instance, ilos.Instance) {
	var err ilos.Instance
	tag, err = Eval(local, tag)
	if err != nil {
		return nil, err
	}
	if ilos.InstanceOf(class.Number, tag) || ilos.InstanceOf(class.Character, tag) {
		return nil, instance.NewDomainError(tag, class.Object)
	}
	object, err = Eval(local, object)
	if err != nil {
		return nil, err
	}
	if _, ok := local.CatchTag.Get(tag); !ok {
		return nil, instance.NewSimpleError(instance.NewString("%v is not defined as the tag"), tag)

	}
	return nil, instance.NewCatchTag(tag, object)
}

func Tagbody(local environment.Environment, body ...ilos.Instance) (ilos.Instance, ilos.Instance) {
	for idx, cadr := range body {
		cddr := instance.GeneralVector(body[idx+1:])
		if !ilos.InstanceOf(class.Cons, cadr) {
			if !local.TagbodyTag.Define(cadr, cddr) {
				return nil, instance.NewImmutableBinding()
			}
		}
	}
	for _, cadr := range body {
		if ilos.InstanceOf(class.Cons, cadr) {
			_, fail := Eval(local, cadr)
			if fail != nil {
			TAG:
				if ilos.InstanceOf(class.TagbodyTag, fail) {
					tag, _ := fail.(instance.Instance).GetSlotValue(instance.NewSymbol("TAG"), class.Escape) // Checked at the top of// This loop
					found := false
					for _, tag1 := range body {
						if tag == tag1 {
							found = true
							break
						}
					}
					if found {
						forms, _ := local.TagbodyTag.Get(tag) // Checked in the function, tagbodyGo
						for _, form := range forms.(instance.GeneralVector) {
							if ilos.InstanceOf(class.Cons, form) {
								_, fail = Eval(local, form)
								if fail != nil {
									goto TAG
								}
							}
						}
						break
					}

				}
				return nil, fail
			}
		}
	}
	return Nil, nil
}

func Go(local environment.Environment, tag ilos.Instance) (ilos.Instance, ilos.Instance) {
	if _, ok := local.TagbodyTag.Get(tag); !ok {
		return nil, instance.NewSimpleError(instance.NewString("%v is not defined as the tag"), tag)

	}
	return nil, instance.NewTagbodyTag(tag)
}

// UnwindProtect first evaluates form. Evaluation of the cleanup-forms always
// occurs, regardless of whether the exit is normal or non-local.
//
// If the form exits normally yielding a value R, then if all of the
// cleanup-forms exit normally the value R is returned by the
// unwind-protect form.
//
// If a non-local exit from form occurs, then the cleanup-forms are executed as
// part of that exit, and then if all of the cleanup-forms exit normally the
// original non-local exit continues.
//
// The cleanup-forms are evaluated from left to right, discarding the resulting
// values. If execution of the cleanup-forms finishes normally, exit from the
// unwind-protect form proceeds as described above. It is permissible for a
// cleanup-form to contain a non-local exit from the unwind-protect form,
// subject to the following constraint:
//
// An error shall be signaled if during execution of the cleanup-forms of an
// unwind-protect form, a non-local exit is executed to a destination which has
// been marked as invalid due to some other non-local exit that is already in
// progress (error-id. control-error).
//
// Note: Because ISLISP does not specify an interactive debugger, it is
// unspecified whether or how error recovery can occur interactively if
// programmatic handling fails. The intent is that if the ISLISP processor does
// not terminate abnormally, normal mechanisms for non-local exit (return-from,
// throw, or go) would be used as necessary and would respect these
// cleanup-forms.
func UnwindProtect(local environment.Environment, form ilos.Instance, cleanupForms ...ilos.Instance) (ilos.Instance, ilos.Instance) {
	ret1, err1 := Eval(local, form)
	ret2, err2 := Progn(local, cleanupForms...)
	if ilos.InstanceOf(class.Escape, err2) {
		return nil, instance.NewControlError()
	}
	if err2 != nil {
		return ret2, err2
	}
	return ret1, err1
}
