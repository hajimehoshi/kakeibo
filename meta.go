package main

import (
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
	// TODO: Need user info?
	ID          UUID     `json:"id,string"`
	LastUpdated UnixTime `json:"last_updated,string"`
	IsDeleted   bool     `json:"is_deleted"`
}

func NewMeta() Meta {
	return Meta{
		ID: GenerateUUID(),
	}
}

func (m *Meta) Reset() {
	m.LastUpdated = 0
}
