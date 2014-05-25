package items

import (
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/models"
)

type Item struct {
	data    *models.ItemData
	view    ItemsView
	storage Storage
}

func NewItem(view ItemsView, storage Storage) *Item {
	return &Item{
		data: &models.ItemData{
			Meta: models.NewMeta(),
			Date: date.Today(),
		},
		view:    view,
		storage: storage,
	}
}

func (i *Item) updateDate(date date.Date) {
	i.data.Date = date
	i.print()
}

func (i *Item) updateSubject(subject string) {
	i.data.Subject = subject
	i.print()
}

func (i *Item) updateAmount(amount models.MoneyAmount) {
	i.data.Amount = amount
	i.print()
}

func (i *Item) print() {
	if i.view == nil {
		return
	}
	i.view.PrintItem(*i.data)
}

func (i *Item) save() error {
	i.data.Meta.LastUpdated = models.UnixTime(0)
	i.print()
	if i.storage == nil {
		return nil
	}
	return i.storage.Save(i.data)
}
