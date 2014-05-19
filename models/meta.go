package models

import (
	"github.com/hajimehoshi/kakeibo/uuid"
	"time"
)

const format = "2006-01-02T15:04:05Z"

type UnixTime int64

func (u UnixTime) MarshalText() ([]byte, error) {
	return []byte(time.Unix(int64(u), 0).UTC().Format(format)), nil
}

func (u *UnixTime) UnmarshalText(s []byte) error {
	t, err := time.Parse(format, string(s))
	if err != nil {
		return err
	}
	*u = UnixTime(t.Unix())
	return nil
}

func (u UnixTime) String() string {
	b, _ := u.MarshalText()
	return string(b)
}

type Meta struct {
	ID          uuid.UUID `json:"id"`
	LastUpdated UnixTime  `json:"last_updated"`
	IsDeleted   bool      `json:"is_deleted"`
	UserID      string    `json:"-"`
}

func NewMeta() Meta {
	return Meta{
		ID: uuid.Generate(),
	}
}

func (m *Meta) Reset() {
	m.LastUpdated = 0
}
