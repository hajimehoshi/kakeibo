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

func (m *Meta) IsValid() bool {
	return m.ID.IsValid()
}
