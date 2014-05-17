package main

import (
	"github.com/hajimehoshi/kakeibo/uuid"
	"strconv"
)

type UnixTime int64

func ParseUnixTime(str string) (UnixTime, error) {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return UnixTime(i), nil
}

func (t UnixTime) String() string {
	return strconv.FormatInt(int64(t), 10)
}

type Meta struct {
	ID          uuid.UUID `json:"id"`
	LastUpdated UnixTime  `json:"last_updated,string"`
	IsDeleted   bool      `json:"is_deleted"`
}

func NewMeta() Meta {
	return Meta{
		ID: uuid.Generate(),
	}
}

func (m *Meta) Reset() {
	m.LastUpdated = 0
}
