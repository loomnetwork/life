package gowasm

import (
	"fmt"
	"github.com/perlin-network/life/exec"
	"syscall"
)

type jsObject map[string]interface{}

func (o jsObject) Get(v string) interface{} {
	return o[v]
}

func (o jsObject) New(vm *exec.VirtualMachine, args ...Value) interface{} {
	return jsObject(make(map[string]interface{}))
}

func (o jsObject) Call(method string, args ...Value) interface{} {
	if m, ok := o[method]; ok {
		if m, ok := m.(func(args ...Value) interface{}); ok {
			return m(args...)
		}
	}
	return newJSError(fmt.Errorf("can not call method %s on jsObject(%#v) ", method, o))
}

type jsError struct {
	jsObject
}

func (e *jsError) Error() string {
	return e.Get("message").(string)
}

func newJSError(err error) jsError {
	jsError := jsError{
		jsObject: jsObject{"message": err.Error()},
	}
	switch e := err.(type) {
	case syscall.Errno:
		jsError.jsObject["code"] = codeByErrno[e]
	}
	return jsError
}
