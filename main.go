// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/idb"
	"github.com/hajimehoshi/kakeibo/models"
	"reflect"
)

var schemaSet = idb.NewSchemaSet()

func init() {
	schemaSet.Add(reflect.TypeOf(&models.ItemData{}), &idb.Schema{
		Name: "items",
	})
}

var items *models.Items

func printError(val interface{}) {
	js.Global.Get("console").Call("error", val)
}

func addEventListeners(form js.Object) {
	inputDate := form.Call("querySelector", "input[name=date]")
	inputDate.Set("onchange", func(e js.Object) {
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
	inputSubject.Set("onchange", func(e js.Object) {
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
	inputMoneyAmount.Set("onchange", func(e js.Object) {
		id, err := getIDFromElement(e.Get("target"))
		if err != nil {
			printError(err.Error())
			return
		}
		item := items.Get(id)

		amount := e.Get("target").Get("value").Int()
		item.UpdateAmount(models.MoneyAmount(amount))
	})
}

func deleteDBIfUserChanged(name string) {
	ls := js.Global.Get("localStorage")
	last := ls.Call("getItem", "last_user_email").Str()
	current := js.Global.Get("userEmail").Str()
	if last != current {
		js.Global.Get("indexedDB").Call("deleteDatabase", name)
		ls.Call("setItem", "last_user_email", current)
	}
}

func main() {
	const dbName = "kakeibo"
	deleteDBIfUserChanged(dbName)

	var view = &HTMLView{}
	db := idb.New(dbName, schemaSet)

	items = models.NewItems(view, db)
	items.Sync()
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
