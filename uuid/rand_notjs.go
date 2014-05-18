// +build !js

package uuid

import (
	crand "crypto/rand"
)

func rand() [16]byte {
	result := [16]byte{}
	crand.Read(result[:])
	return result
}
