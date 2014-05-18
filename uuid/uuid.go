package uuid

import (
	//"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const length = 16

const Zero = UUID("00000000-0000-4000-8000-000000000000")

// Google App Engine's datastore doesn't accept a fixed-size array. Use string
// instead...
type UUID string

func ParseString(str string) (UUID, error) {
	const pattern = "^([0-9a-fA-F]{8})-([0-9a-fA-F]{4})-" +
		"(4[0-9a-fA-F]{3})-([89abAB][0-9a-fA-F]{3})-([0-9a-fA-F]{12})$"
	m, err := regexp.MatchString(pattern, str)
	if err != nil {
		return Zero, err
	}
	if !m {
		return Zero, errors.New("uuid: invalid UUID (v4) string")
	}
	return UUID(strings.ToLower(str)), nil
}

func (i UUID) String() string {
	return string(i)
}

func (i UUID) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *UUID) UnmarshalText(text []byte) (err error) {
	*i, err = ParseString(string(text))
	return
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
