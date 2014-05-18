package main

import (
	"encoding/json"
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/models"
	"github.com/hajimehoshi/kakeibo/uuid"
)

type Storage interface {
	Save(interface{}) error
	Load(name string, id uuid.UUID, callback func(val string)) error
	LoadAll(name string, callback func(val []string)) error
}

type ItemView interface {
	PrintItem(data models.ItemData)
	SetIDsToItemTable(ids []uuid.UUID)
}

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

func (i *Item) Save() {
	i.data.Meta.LastUpdated = 0
	i.Print()
	i.save()
}

func (i *Item) Destroy() {
	meta := i.data.Meta
	meta.LastUpdated = 0
	meta.IsDeleted = true
	i.data = &models.ItemData{Meta: meta}
	i.Save()
}

func (i *Item) Print() {
	if i.view == nil {
		return
	}
	i.view.PrintItem(*i.data)
}

func (i *Item) save() {
	if i.storage == nil {
		return
	}
	err := i.storage.Save(i.data)
	if err != nil {
		print(err.Error())
	}
}

type Items struct {
	items   map[uuid.UUID]*Item
	view    ItemView
	storage Storage
}

func NewItems(view ItemView, storage Storage) *Items {
	items := &Items{
		items:   map[uuid.UUID]*Item{},
		view:    view,
		storage: storage,
	}
	items.storage.LoadAll("items", items.onStorageItemsLoaded)
	return items
}

func toItemData(val string) *models.ItemData {
	var d models.ItemData
	if err := json.Unmarshal([]byte(val), &d); err != nil {
		print(err.Error())
		return nil
	}
	return &d
}

func (i *Items) onStorageItemsLoaded(vals []string) {
	ids := []uuid.UUID{}
	for _, v := range vals {
		d := toItemData(v)
		i.onItemLoaded(d)
		ids = append(ids, d.Meta.ID)
	}
	i.view.SetIDsToItemTable(ids)
	for _, id := range ids {
		item := i.items[id]
		item.Print()
	}
}

func (i *Items) onStorageItemLoaded(val string) {
	i.onItemLoaded(toItemData(val))
}

func (i *Items) onItemLoaded(d *models.ItemData) {
	id := d.Meta.ID
	if item, ok := i.items[id]; ok {
		*item.data = *d
		item.Print()
		return
	}
	item := &Item{
		data:    d,
		view:    i.view,
		storage: i.storage,
	}
	i.items[id] = item
	item.Print()
}

func (i *Items) New() *Item {
	item := NewItem(i.view, i.storage)
	i.items[item.data.Meta.ID] = item
	return item
}

func (i *Items) Get(id uuid.UUID) *Item {
	if item, ok := i.items[id]; ok {
		return item
	}
	return nil
}

func (i *Items) GetItems(ids []uuid.UUID) []*Item {
	items := []*Item{}
	for _, id := range ids {
		item := i.Get(id)
		if item != nil {
			items = append(items, item)
		}
	}
	return items
}

func (i *Items) All() []*Item {
	result := []*Item{}
	for _, item := range i.items {
		result = append(result, item)
	}
	return result
}
