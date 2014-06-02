package models

import (
	"github.com/hajimehoshi/kakeibo/uuid"
	"time"
)

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

func (m *Meta) IsValid() bool {
	return m.ID.IsValid()
}
