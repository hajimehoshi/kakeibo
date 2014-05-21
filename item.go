package main

import (
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/models"
	"github.com/hajimehoshi/kakeibo/uuid"
	"reflect"
)

type Storage interface {
	Save(interface{}) error
}

type ItemView interface {
	PrintItem(data models.ItemData)
	OnInit(items *Items)
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
	i.data.Meta.LastUpdated = models.UnixTime(0)
	i.Print()
	i.save()
}

func (i *Item) Destroy() {
	meta := i.data.Meta
	meta.LastUpdated = models.UnixTime(0)
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
	return &Items{
		items:   map[uuid.UUID]*Item{},
		view:    view,
		storage: storage,
	}
}

func (i *Items) Type() reflect.Type {
	return reflect.TypeOf((*models.ItemData)(nil)).Elem()
}

func (i *Items) OnLoaded(vals []interface{}) {
	for _, v := range vals {
		d, ok := v.(*models.ItemData)
		if !ok {
			print("invalid data")
			return
		}
		id := d.Meta.ID
		if item, ok := i.items[id]; ok {
			*item.data = *d
			item.Print()
			continue
		}
		item := &Item{
			data:    d,
			view:    i.view,
			storage: i.storage,
		}
		i.items[id] = item
		item.Print()
	}
}

func (i *Items) OnInitialLoaded(vals []interface{}) {
	i.OnLoaded(vals)
	i.view.OnInit(i)
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

func (i *Items) GetAll() []*Item {
	result := []*Item{}
	for _, item := range i.items {
		result = append(result, item)
	}
	return result
}
