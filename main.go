// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/idb"
	"github.com/hajimehoshi/kakeibo/models"
	"time"
)

var items *Items

func printError(val interface{}) {
	js.Global.Get("console").Call("error", val)
}

func addEventListeners(form js.Object) {
	inputDate := form.Call("querySelector", "input[name=Date]")
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
	inputSubject := form.Call("querySelector", "input[name=Subject]")
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
	inputMoneyAmount := form.Call("querySelector", "input[name=Amount]")
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

func deleteDBIfUserChanged(name string, callback func()) {
	ls := js.Global.Get("localStorage")
	last := ls.Call("getItem", "last_user_email").Str()
	current := js.Global.Get("userEmail").Str()
	if last == current {
		callback()
		return
	}
	req := js.Global.Get("indexedDB").Call("deleteDatabase", name)
	req.Set("onsuccess", callback)
	ls.Call("setItem", "last_user_email", current)
}

const dbName = "kakeibo"

func main() {
	deleteDBIfUserChanged(dbName, ready)
}

var view = NewHTMLView()

func ready() {
	db := idb.New(dbName)

	document := js.Global.Get("document")

	items = NewItems(view, view, db)
	item := items.New()
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

		newItem := items.New()
		form.Get("dataset").Set(datasetAttrID, newItem.ID().String())
		newItem.Print()

		view.AddIDToItemTable(item.ID())
		items.Get(id).Print()
	})
	form.Get("dataset").Set(datasetAttrID, item.ID().String())
	item.Print()

	var sync func()
	sync = func() {
		db.Sync([]idb.Model{items})
		time.AfterFunc(60 * time.Second, sync)
	}
	sync()

	debugLink := document.Call("getElementById", "debug_link")
	debugLink.Set("onclick", toggleDebugOverlay)

	debugOverlay := document.Call("getElementById", "debug_overlay")
	debugOverlay.Set("onclick", toggleDebugOverlay)

	js.Global.Get("window").Set("onhashchange", onHashChange)
	js.Global.Get("window").Call("onhashchange")
}

func toggleDebugOverlay(e js.Object) {
	e.Call("preventDefault")
	d := js.Global.Get("document").Call("getElementById", "debug_overlay")
	if d.Get("style").Get("display").Str() == "block" {
		d.Get("style").Set("display", "none")
		return
	}
	d.Get("style").Set("display", "block")
}

func onHashChange(e js.Object) {
	hash := js.Global.Get("location").Get("hash").Str()
	// Remove the initial '#'
	if 1 <= len(hash) {
		hash = hash[1:]
	}
	if hash == "" {
		href := js.Global.Get("location").Get("href").Str()
		if 0 < len(href) && href[len(href)-1] == '#' {
			href = href[:len(href)-2]
			js.Global.Get("history").Call(
				"replaceState", "", "", href)
		}
		view.UpdateMode(ViewModeAll, date.Date(0))
		return
	}
	ym, err := date.ParseISO8601(hash + "-01")
	if err != nil {
		printError(err.Error())
		return
	}
	view.UpdateMode(ViewModeYearMonth, ym)
}
