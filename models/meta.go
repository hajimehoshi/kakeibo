package models

import (
	"github.com/hajimehoshi/kakeibo/uuid"
)

type UnixTime float64

type Meta struct {
	ID          uuid.UUID `json:"id" datastore:",string"`
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
