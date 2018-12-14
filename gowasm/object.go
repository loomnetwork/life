package gowasm

import (
	"fmt"
	"github.com/perlin-network/life/exec"
	"syscall"
)

type JSObject map[string]interface{}

func (o JSObject) Get(v string) interface{} {
	return o[v]
}

func (o JSObject) New(vm *exec.VirtualMachine, args ...Value) interface{} {
	return JSObject(make(map[string]interface{}))
}

func (o JSObject) Call(method string, args ...Value) interface{} {
	if m, ok := o[method]; ok {
		if m, ok := m.(func(args ...Value) interface{}); ok {
			return m(args...)
		}
	}
	return NewJSError(fmt.Errorf("can not call method %s on JSObject(%#v) ", method, o))
}

type jsError struct {
	JSObject
}

func (e *jsError) Error() string {
	return e.Get("message").(string)
}

func NewJSError(err error) jsError {
	jsError := jsError{
		JSObject: JSObject{"message": err.Error()},
	}
	switch e := err.(type) {
	case syscall.Errno:
		jsError.JSObject["code"] = codeByErrno[e]
	}
	return jsError
}

