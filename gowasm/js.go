// +build js,wasm

package gowasm

import (
	"syscall/js"
)

func JSValue2ByteSlice(data js.Value) []byte {
	bs := make([]byte, data.Length())
	for i := range bs {
		bs[i] = byte(data.Index(i).Int())
	}
	return bs
}

func ByteSlice2JSValue(data []byte) []interface{} {
	is := make([]interface{}, len(data))
	for i, d := range data {
		is[i] = d
	}
	return is
}