// +build js

package main

import (
	"errors"
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/idb"
	"github.com/hajimehoshi/kakeibo/uuid"
)

var schemaSet = idb.NewSchemaSet()

var items *Items

func printError(val interface{}) {
	js.Global.Get("console").Call("error", val)
}

func getIDFromElement(e js.Object) (uuid.UUID, error) {
	for {
		attr := e.Get("dataset").Get(datasetAttrID)
		if !attr.IsUndefined() {
			str := attr.Str()
			id, err := uuid.ParseString(str)
			if err != nil {
				return uuid.UUID{}, err
			}
			return id, nil
		}
		e = e.Get("parentNode")
		if e.IsNull() || e.IsUndefined() {
			break
		}
	}
	return uuid.UUID{}, errors.New("not found")
}

func addEventListeners(form js.Object) {
	inputDate := form.Call("querySelector", "input[name=date]")
	inputDate.Call("addEventListener", "change", func(e js.Object) {
		id, err := getIDFromElement(e.Get("target"))
		if err != nil {
			printError(err.Error())
			return
		}
		item := items.Get(id)

		dateStr := e.Get("target").Get("value").Str()
		d, err := date.ParseISO8601(dateStr)
		if err != nil {
			printError(err.Error())
			return
		}
		item.UpdateDate(d)
	})
	inputSubject := form.Call("querySelector", "input[name=subject]")
	inputSubject.Call("addEventListener", "change", func(e js.Object) {
		id, err := getIDFromElement(e.Get("target"))
		if err != nil {
			printError(err.Error())
			return
		}
		item := items.Get(id)

		subject := e.Get("target").Get("value").Str()
		item.UpdateSubject(subject)
	})
	inputMoneyAmount := form.Call("querySelector", "input[name=amount]")
	inputMoneyAmount.Call("addEventListener", "change", func(e js.Object) {
		id, err := getIDFromElement(e.Get("target"))
		if err != nil {
			printError(err.Error())
			return
		}
		item := items.Get(id)

		amount := e.Get("target").Get("value").Int()
		item.UpdateAmount(MoneyAmount(amount))
	})
}

func main() {
	var view = &HTMLView{}
	db := idb.New("kakeibo", schemaSet)

	items = NewItems(view, db)
	item := items.New()
	document := js.Global.Get("document")
	form := document.Call("getElementById", "form_item")
	addEventListeners(form)
	form.Set("onsubmit", func(e js.Object) {
		e.Call("preventDefault")
		form := e.Get("target")
		id, err := getIDFromElement(form)
		if err != nil {
			printError(err.Error())
			return
		}
		item := items.Get(id)
		// TODO: validation here?
		item.Save()
		view.AddIDToItemTable(item.ID())
		items.Get(id).Print()

		item = items.New()
		form.Get("dataset").Set(datasetAttrID, item.ID().String())
		item.Print()
	})
	form.Get("dataset").Set(datasetAttrID, item.ID().String())
	item.Print()
}
