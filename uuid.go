// +build js

package main

import (
	"fmt"
	"github.com/gopherjs/gopherjs/js"
)

type UUID [16]byte

func (i UUID) String() string {
	return fmt.Sprintf(
		"%x-%x-%x-%x-%x", i[0:4], i[4:6], i[6:8], i[8:10], i[10:])
}

// rand generates random byte sequence.
func rand() [16]byte {
	result := [16]byte{}
	js.Global.Get("crypto").Call("getRandomValues", result[:])
	return result
}

func GenerateUUID() UUID {
	id := UUID{}
	r := rand()
	copy(id[:], r[:])
	const version = 4
	id[8] = (id[8] & 0x3f) | 0x80
	id[6] = (id[6] & 0x0f) | (version << 4)
	return id
}
