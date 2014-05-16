package main

import (
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/idb"
	"reflect"
)

func init() {
	schemaSet.Add(reflect.TypeOf(&ItemData{}), &idb.Schema{
		Name: "items",
	})
}

type Saver interface {
	Save(interface{}) error
}

type Loader interface {
	Load(id UUID) interface{}
}

type MoneyAmount int

type ItemData struct {
	Meta    Meta        `json:"meta"`
	Date    date.Date   `json:"date"`
	Subject string      `json:"subject"`
	Amount  MoneyAmount `json:"amount"`
}

type ItemPrinter interface {
	PrintDate(date date.Date)
	PrintSubject(subject string)
	PrintMoneyAmount(amount MoneyAmount)
}

type Item struct {
	data    *ItemData
	printer ItemPrinter
	saver   Saver
}

func NewItem(saver Saver) *Item {
	// TODO: Use cache and load from here. Don't create two instances with
	// the same ID.
	item := &Item{
		data: &ItemData{
			Meta: NewMeta(),
			Date: date.Today(),
		},
		saver: saver,
	}
	return item
}

func (i *Item) SetPrinter(printer ItemPrinter) {
	i.printer = printer
	i.print()
}

func (i *Item) UpdateDate(date date.Date) {
	i.data.Date = date
	i.print()
}

func (i *Item) UpdateSubject(subject string) {
	i.data.Subject = subject
	i.print()
}

func (i *Item) UpdateAmount(amount MoneyAmount) {
	i.data.Amount = amount
	i.print()
}

func (i *Item) Save() {
	i.data.Meta.LastUpdated = 0
	i.print()
	i.save()
}

func (i *Item) Destroy() {
	meta := i.data.Meta
	meta.LastUpdated = 0
	meta.IsDeleted = true
	i.data = &ItemData{Meta: meta}
	i.Save()
}

func (i *Item) print() {
	if i.printer == nil {
		return
	}
	i.printer.PrintDate(i.data.Date)
	i.printer.PrintSubject(i.data.Subject)
	i.printer.PrintMoneyAmount(i.data.Amount)
}

func (i *Item) save() {
	if i.saver == nil {
		return
	}
	err := i.saver.Save(i.data)
	if err != nil {
		print(err.Error())
	}
}

type Items struct {
	// TODO: Revert the key to UUID
	items  map[string]*Item
	saver  Saver
	loader Loader
}

func NewItems(saver Saver, loader Loader) *Items {
	return &Items{
		items:  map[string]*Item{},
		saver:  saver,
		loader: loader,
	}
}

func (i *Items) New() *Item {
	item := NewItem(i.saver)
	i.items[item.data.Meta.ID.String()] = item
	return item
}

func (i *Items) Get(id UUID) *Item {
	if item, ok := i.items[id.String()]; ok {
		return item
	}
	data, ok := i.loader.Load(id).(*ItemData)
	if !ok {
		return nil
	}
	if data.Meta.ID != id {
		panic("invalid data")
		return nil
	}
	item := &Item{
		data:  data,
		saver: i.saver,
	}
	i.items[id.String()] = item
	return item
}
