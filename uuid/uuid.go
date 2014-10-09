package uuid

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const parsePattern = "^([0-9a-fA-F]{8})-([0-9a-fA-F]{4})-" +
	"(4[0-9a-fA-F]{3})-([89abAB][0-9a-fA-F]{3})-([0-9a-fA-F]{12})$"

var parseReg = regexp.MustCompile(parsePattern)

const strictPattern = "^([0-9a-f]{8})-([0-9a-f]{4})-" +
	"(4[0-9a-f]{3})-([89ab][0-9a-f]{3})-([0-9a-f]{12})$"

var strictReg = regexp.MustCompile(strictPattern)

// Google App Engine's datastore doesn't accept a fixed-size array. Use string
// instead...
type UUID string

func ParseString(str string) (UUID, error) {
	if !parseReg.MatchString(str) {
		return *new(UUID), errors.New("uuid: invalid UUID (v4) string")
	}
	return UUID(strings.ToLower(str)), nil
}

func (i UUID) String() string {
	return string(i)
}

func (i UUID) MarshalText() ([]byte, error) {
	if !i.IsValid() {
		return nil, errors.New("UUID.MarshalText: invalid UUID format")
	}
	return []byte(i.String()), nil
}

func (i *UUID) UnmarshalText(text []byte) (err error) {
	*i, err = ParseString(string(text))
	return
}

func (i UUID) IsValid() bool {
	return strictReg.MatchString(string(i))
}

func Generate() UUID {
	id := [16]byte{}
	r := rand()
	copy(id[:], r[:])
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	str := fmt.Sprintf(
		"%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:])
	return UUID(str)
}
