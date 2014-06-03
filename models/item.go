package models

import (
	"github.com/hajimehoshi/kakeibo/date"
	"time"
)

// FIXME: Implement 'income' and 'outgo'

type ItemData struct {
	Meta    Meta
	Date    date.Date
	Subject string
	Amount  int32
}

func (i *ItemData) IsValid() bool {
	if !i.Meta.IsValid() {
		return false
	}
	if i.Meta.IsDeleted {
		return true
	}
	if i.Subject == "" {
		return false
	}
	return true
}

func (i *ItemData) Destroy() {
	meta := i.Meta
	meta.IsDeleted = true
	*i = ItemData{Meta: meta}
}








