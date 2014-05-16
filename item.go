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

func NewItem() *Item {
	// TODO: Use cache and load from here. Don't create two instances with the same ID.
	item := &Item{
		data: &ItemData{
			Meta: NewMeta(),
			Date: date.Today(),
		},
	}
	return item
}

func (i *Item) SetPrinter(printer ItemPrinter) {
	i.printer = printer
	i.print()
}

func (i *Item) SetSaver(saver Saver) {
	i.saver = saver
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
