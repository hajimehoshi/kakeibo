package models

import (
	"github.com/hajimehoshi/kakeibo/date"
)

// FIXME: Implement 'income' and 'outgo'

type ItemData struct {
	Meta    Meta
	Date    date.Date
	Subject string
	Amount  int32
}

func (i *ItemData) IsValid() bool {
	if i.Subject == "" {
		return false
	}
	return true
}
