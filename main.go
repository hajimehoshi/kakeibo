package main

import (
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/nu7hatch/gouuid"
	"time"
)

type Date struct {
	t time.Time
}

func NewDate(year int, month time.Month, day int) Date {
	return Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

func Today() Date {
	return Date{time.Now().UTC()}
}

func (d Date) AddDate(years, months, days int) Date {
	return Date{d.t.AddDate(years, months, days)}
}

func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.t.Year(), d.t.Month(), d.t.Day())
}

type MoneyAmount int

type ItemPrinter interface {
	PrintDate(date Date)
	PrintSubject(subject string)
	PrintMoneyAmount(amount MoneyAmount)
}

type Item struct {
	id      uuid.UUID
	date    Date
	subject string
	amount  MoneyAmount
	printer ItemPrinter
}

func NewItem(printer ItemPrinter) *Item {
	/*id, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}*/
	item := &Item{
		date:    Today(),
		printer: printer,
	}
	item.print()
	return item
}

func (i *Item) UpdateDate(date Date) {
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

type FormItemPrinter struct {
	form js.Object
}

func (f *FormItemPrinter) PrintDate(date Date) {
	input := f.form.Call("querySelector", "input[name=date]")
	input.Set("value", date.String())
}

func (f *FormItemPrinter) PrintSubject(subject string) {
	input := f.form.Call("querySelector", "input[name=subject]")
	input.Set("value", subject)
}

func (f *FormItemPrinter) PrintMoneyAmount(amount MoneyAmount) {
	input := f.form.Call("querySelector", "input[name=money_amount]")
	input.Set("value", amount)
}

func main() {
	document := js.Global.Get("document")
	form := document.Call("getElementById", "form_record")
	printer := &FormItemPrinter{form}
	item := NewItem(printer)
	print(item)
	// TODO: Move this call somewhere
	//js.Global.Call("alert", "Hello, Kakeibo!")
}
