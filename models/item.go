package models

import (
	"github.com/hajimehoshi/kakeibo/date"
)

type MoneyAmount int

type ItemData struct {
	Meta    Meta        `json:"meta"`
	Date    date.Date   `json:"date"`
	Subject string      `json:"subject"`
	Amount  MoneyAmount `json:"amount"`
}

func (i *ItemData) IsValid() bool {
	if i.Subject == "" {
		return false
	}
	return true
}
