package items

import (
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/models"
	"time"
)

type Item struct {
	data    *models.ItemData
	view    ItemsView
	storage Storage
}

func newItem(view ItemsView, storage Storage) *Item {
	return &Item{
		data: &models.ItemData{
			Meta: models.NewMeta(),
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

func (i *Item) updateAmount(amount int32) {
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
	i.data.Meta.LastUpdated = time.Time{}
	i.print()
	if i.storage == nil {
		return nil
	}
	return i.storage.Save(i.data)
}
