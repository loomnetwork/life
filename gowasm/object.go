package gowasm

import "github.com/perlin-network/life/exec"

type jsObject map[string]interface{}

func (o jsObject) Get(v string) interface{} {
	return o[v]
}

func (o jsObject) New(vm *exec.VirtualMachine,args ...interface{})  interface{} {
	return jsObject(make(map[string]interface{}))
}

func (o jsObject) Call(method string, args ...interface{}) interface{} {
	if m, ok := o[method]; ok {
		if m, ok := m.(func(args ...interface{}) interface{}); ok {
			return m(args...)
		}
	}
	return nil
}

func newJSError(err string) jsObject {
	return jsObject{"message": err}
}