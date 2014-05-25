package items

import (
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/models"
	"github.com/hajimehoshi/kakeibo/uuid"
)

type Item struct {
	data    *models.ItemData
	view    ItemView
	storage Storage
}

func NewItem(view ItemView, storage Storage) *Item {
	return &Item{
		data: &models.ItemData{
			Meta: models.NewMeta(),
			Date: date.Today(),
		},
		view:    view,
		storage: storage,
	}
}

// TODO: Remove this getter
func (i *Item) ID() uuid.UUID {
	return i.data.Meta.ID
}

func (i *Item) UpdateDate(date date.Date) {
	i.data.Date = date
	i.Print()
}

func (i *Item) UpdateSubject(subject string) {
	i.data.Subject = subject
	i.Print()
}

func (i *Item) UpdateAmount(amount models.MoneyAmount) {
	i.data.Amount = amount
	i.Print()
}

func (i *Item) Destroy() {
	meta := i.data.Meta
	meta.LastUpdated = models.UnixTime(0)
	meta.IsDeleted = true
	i.data = &models.ItemData{Meta: meta}
	i.save()
}

func (i *Item) Print() {
	if i.view == nil {
		return
	}
	i.view.PrintItem(*i.data)
}

func (i *Item) save() error {
	i.data.Meta.LastUpdated = models.UnixTime(0)
	i.Print()
	if i.storage == nil {
		return nil
	}
	return i.storage.Save(i.data)
}
