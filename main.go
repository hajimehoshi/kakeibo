// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/date"
)

var schemaSet = NewSchemaSet()

var db *IDB

type Saver interface {
	Save(value interface{}) error
}

func printError(val interface{}) {
	js.Global.Get("console").Call("error", val)
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
			printError(err.Error())
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

type IDBObserverImpl struct {}

func (t *IDBObserverImpl) OnReady(d *IDB) {
	ready()
}

func main() {
	db = NewIDB("kakeibo", schemaSet, &IDBObserverImpl{})
}

func ready() {
	document := js.Global.Get("document")
	form := document.Call("getElementById", "form_record")
	item := NewItem()
	printer := NewItemForm(item, form)
	_ = printer

	item.SetSaver(db)
	item.Save()
}
