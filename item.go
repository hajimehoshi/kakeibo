package main

import (
	"encoding/json"
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/idb"
	"strconv"
	"reflect"
)

func init() {
	schemaSet.Add(reflect.TypeOf(&Item{}), &idb.Schema{
		Name: "items",
	})
}

type MoneyAmount int

type ItemPrinter interface {
	PrintDate(date date.Date)
	PrintSubject(subject string)
	PrintMoneyAmount(amount MoneyAmount)
}

type Item struct {
	meta    Meta        `json:"meta"`
	date    date.Date   `json:"date"`
	subject string      `json:"subject"`
	amount  MoneyAmount `json:"amount"`
	printer ItemPrinter `json:"-"`
	saver   Saver       `json:"-"`
}

func NewItem() *Item {
	item := &Item{
		meta: NewMeta(),
		date: date.Today(),
	}
	return item
}

func (i *Item) MarshalJSON() ([]byte, error) {
	// TODO: Use reflect
	m := map[string]interface{}{}
	m["id"] = i.meta.ID.String()
	m["last_updated"] = strconv.FormatInt(i.meta.LastUpdated, 10)
	m["is_deleted"] = i.meta.IsDeleted
	m["date"] = i.date.String()
	m["subject"] = i.subject
	m["amount"] = int(i.amount)
	return json.Marshal(m)
}

func (i *Item) SetPrinter(printer ItemPrinter) {
	i.printer = printer
	i.print()
}

func (i *Item) SetSaver(saver Saver) {
	i.saver = saver
}

func (i *Item) UpdateDate(date date.Date) {
	i.date = date
	i.print()
}

func (i *Item) UpdateSubject(subject string) {
	i.subject = subject
	i.print()
}

func (i *Item) UpdateAmount(amount MoneyAmount) {
	i.amount = amount
	i.print()
}

func (i *Item) Save() {
	i.save()
	i.print()
}

func (i *Item) print() {
	if i.printer == nil {
		return
	}
	i.printer.PrintDate(i.date)
	i.printer.PrintSubject(i.subject)
	i.printer.PrintMoneyAmount(i.amount)
}

func (i *Item) save() {
	if i.saver == nil {
		return
	}
	err := i.saver.Save(i)
	if err != nil {
		print(err.Error())
	}
}
