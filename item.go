package main

import (
	"encoding/json"
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/idb"
	"github.com/hajimehoshi/kakeibo/uuid"
	"reflect"
)

func init() {
	schemaSet.Add(reflect.TypeOf(&ItemData{}), &idb.Schema{
		Name: "items",
	})
}

type Storage interface {
	Save(interface{}) error
	Load(name string, id uuid.UUID, callback func(val string)) error
	LoadAll(name string, callback func(val string)) error
}

type MoneyAmount int

type ItemData struct {
	Meta    Meta        `json:"meta"`
	Date    date.Date   `json:"date"`
	Subject string      `json:"subject"`
	Amount  MoneyAmount `json:"amount"`
}

type ItemView interface {
	PrintItem(data ItemData)
}

type Item struct {
	data    *ItemData
	view    ItemView
	storage Storage
}

func NewItem(view ItemView, storage Storage) *Item {
	return &Item{
		data: &ItemData{
			Meta: NewMeta(),
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

func (i *Item) UpdateAmount(amount MoneyAmount) {
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
	i.data = &ItemData{Meta: meta}
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
	storage.LoadAll("items", items.onStorageItemLoaded)
	return items
}

func (i *Items) onStorageItemLoaded(val string) {
	var d ItemData
	if err := json.Unmarshal([]byte(val), &d); err != nil {
		print(err.Error())
		return
	}
	id := d.Meta.ID
	if item, ok := i.items[id]; ok {
		*item.data = d
		item.Print()
		return
	}
	item := &Item{
		data:    &d,
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
	// TODO: Is this necessary?
	i.storage.Load("items", id, i.onStorageItemLoaded)
	return nil
}

func (i *Items) All() []*Item {
	result := []*Item{}
	for _, item := range i.items {
		result = append(result, item)
	}
	return result
}
