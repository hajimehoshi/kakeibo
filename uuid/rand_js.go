// +build js

package uuid

import (
	"github.com/gopherjs/gopherjs/js"
)

// rand generates random byte sequence.
func rand() [16]byte {
	result := [16]byte{}
	js.Global.Get("crypto").Call("getRandomValues", result[:])
	return result
}
