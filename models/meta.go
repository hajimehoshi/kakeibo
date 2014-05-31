package models

import (
	"github.com/hajimehoshi/kakeibo/uuid"
	"time"
)

/*const format = "2006-01-02T15:04:05.000Z"

type Time time.Time

func (u Time) MarshalText() ([]byte, error) {
	return []byte(time.Time(u).UTC().Format(format)), nil
}

func (u *Time) UnmarshalText(s []byte) error {
	t, err := time.Parse(format, string(s))
	if err != nil {
		return err
	}
	*u = Time(t.UTC())
	return nil
}

func (u Time) String() string {
	b, _ := u.MarshalText()
	return string(b)
}

func (u Time) LessThan(v Time) bool {
	return time.Time(u).UnixNano() < time.Time(v).UnixNano()
}*/

type Meta struct {
	ID          uuid.UUID
	LastUpdated time.Time
	IsDeleted   bool
	UserID      string `json:"-"`
}

func NewMeta() Meta {
	return Meta{
		ID: uuid.Generate(),
	}
}
