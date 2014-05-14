// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/date"
)

func printError(val interface{}) {
	js.Global.Get("console").Call("error", val)
}

type MoneyAmount int

type ItemPrinter interface {
	PrintDate(date date.Date)
	PrintSubject(subject string)
	PrintMoneyAmount(amount MoneyAmount)
}

type ItemSaver interface {
	Save(val map[string]interface{})
}

type Item struct {
	id      UUID        `json:"id"`
	date    date.Date   `json:"date"`
	subject string      `json:"subject"`
	amount  MoneyAmount `json:"amount"`
	printer ItemPrinter
}

func NewItem() *Item {
	/*id, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}*/
	item := &Item{
		date: date.Today(),
	}
	return item
}

func (i *Item) SetPrinter(printer ItemPrinter) {
	i.printer = printer
	i.print()
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
	// FIXME: Implement this
	i.print()
}

func (i *Item) print() {
	// TODO: Check ID here?
	i.printer.PrintDate(i.date)
	i.printer.PrintSubject(i.subject)
	i.printer.PrintMoneyAmount(i.amount)
}

type ItemForm struct {
	item *Item
	form js.Object
}

func NewItemForm(item *Item, form js.Object) *ItemForm {
	f := &ItemForm{item, form}
	item.SetPrinter(f)
	f.addEventHandlers()
	return f
}

func (f *ItemForm) addEventHandlers() {
	inputDate := f.form.Call("querySelector", "input[name=date]")
	inputDate.Call("addEventListener", "input", func() {
		dateStr := inputDate.Get("value").Str()
		date, err := date.ParseISO8601(dateStr)
		if err != nil {
			printError(err)
			return
		}
		f.item.UpdateDate(date)
	})
	inputSubject := f.form.Call("querySelector", "input[name=subject]")
	inputSubject.Call("addEventListener", "input", func() {
		print("hoge")
	})
	inputMoneyAmount := f.form.Call("querySelector", "input[name=money_amount]")
	inputMoneyAmount.Call("addEventListener", "input", func() {
		print("hoge")
	})
}

func (f *ItemForm) PrintDate(date date.Date) {
	input := f.form.Call("querySelector", "input[name=date]")
	input.Set("value", date.String())
}

func (f *ItemForm) PrintSubject(subject string) {
	input := f.form.Call("querySelector", "input[name=subject]")
	input.Set("value", subject)
}

func (f *ItemForm) PrintMoneyAmount(amount MoneyAmount) {
	input := f.form.Call("querySelector", "input[name=money_amount]")
	input.Set("value", amount)
}

func main() {
	for i := 0; i < 100; i++ {
		print(GenerateUUID().String())
	}

	document := js.Global.Get("document")
	form := document.Call("getElementById", "form_record")
	item := NewItem()
	printer := NewItemForm(item, form)
	print(printer)
	// TODO: Move this call somewhere
	//js.Global.Call("alert", "Hello, Kakeibo!")
}
