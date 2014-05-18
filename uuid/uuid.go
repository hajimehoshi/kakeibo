package uuid

import (
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type UUID [16]byte

func ParseString(str string) (UUID, error) {
	const pattern = "^([0-9a-fA-F]{8})-([0-9a-fA-F]{4})-" +
		"(4[0-9a-fA-F]{3})-([89abAB][0-9a-fA-F]{3})-([0-9a-fA-F]{12})$"
	re := regexp.MustCompile(pattern)

	match := re.FindStringSubmatch(str)
	if match == nil {
		return UUID{}, errors.New("uuid: invalid UUID string")
	}
	bytes, err := hex.DecodeString(strings.Join(match[1:6], ""))
	if err != nil {
		return UUID{}, err
	}
	id := UUID{}
	copy(id[:], bytes)
	return id, nil
}

func (i UUID) String() string {
	return fmt.Sprintf(
		"%x-%x-%x-%x-%x", i[0:4], i[4:6], i[6:8], i[8:10], i[10:])
}

func (i UUID) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *UUID) UnmarshalText(text []byte) (err error) {
	*i, err = ParseString(string(text))
	return
}

func Generate() UUID {
	id := UUID{}
	r := rand()
	copy(id[:], r[:])
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	return id
}
