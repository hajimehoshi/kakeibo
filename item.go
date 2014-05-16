package main

import (
	//"encoding/json"
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/idb"
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
	Meta    Meta        `json:"meta"`
	Date    date.Date   `json:"date"`
	Subject string      `json:"subject"`
	Amount  MoneyAmount `json:"amount"`
	printer ItemPrinter `json:"-"`
	saver   Saver       `json:"-"`
}

func NewItem() *Item {
	// TODO: Use cache and load from here. Don't create two instances with the same ID.
	item := &Item{
		Meta: NewMeta(),
		Date: date.Today(),
	}
	return item
}

/*func (i *Item) MarshalJSON() ([]byte, error) {
	// TODO: Use reflect
	m := map[string]interface{}{}
	var err error
	m["meta"], err = json.Marshal(i.meta)
	if err != nil {
		return nil, err
	}
	m["date"] = i.date.String()
	m["subject"] = i.subject
	m["amount"] = int(i.amount)
	return json.Marshal(m)
}*/

func (i *Item) SetPrinter(printer ItemPrinter) {
	i.printer = printer
	i.print()
}

func (i *Item) SetSaver(saver Saver) {
	i.saver = saver
}

func (i *Item) UpdateDate(date date.Date) {
	i.Date = date
	i.print()
}

func (i *Item) UpdateSubject(subject string) {
	i.Subject = subject
	i.print()
}

func (i *Item) UpdateAmount(amount MoneyAmount) {
	i.Amount = amount
	i.print()
}

func (i *Item) Save() {
	i.Meta.LastUpdated = 0
	i.print()
	i.save()
}

func (i *Item) Destroy() {
	meta := i.Meta
	meta.IsDeleted = true
	*i = Item{Meta: meta}
	i.Save()
}

func (i *Item) print() {
	if i.printer == nil {
		return
	}
	i.printer.PrintDate(i.Date)
	i.printer.PrintSubject(i.Subject)
	i.printer.PrintMoneyAmount(i.Amount)
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
