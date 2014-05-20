package models

import (
	"github.com/hajimehoshi/kakeibo/date"
)

type MoneyAmount int

type ItemData struct {
	Meta    Meta
	Date    date.Date
	Subject string
	Amount  MoneyAmount
}

func (i *ItemData) IsValid() bool {
	if i.Subject == "" {
		return false
	}
	return true
}
