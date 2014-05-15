package main

import (
	"math"
)

type Meta struct {
	ID          UUID  `json:"id"`
	LastUpdated int64 `json:"last_updated,string"`
	IsDeleted   bool  `json:"is_deleted"`
}

const NotUpdated int64 = math.MinInt64

func NewMeta() Meta {
	return Meta{
		ID:          GenerateUUID(),
		LastUpdated: NotUpdated,
		IsDeleted:   false,
	}
}

func (m *Meta) Reset() {
	m.LastUpdated = 0
}
