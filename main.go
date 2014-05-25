// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/idb"
	"github.com/hajimehoshi/kakeibo/items"
	"time"
)

func printError(val interface{}) {
	js.Global.Get("console").Call("error", val)
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

var view *HTMLView

func ready() {
	db := idb.New(dbName)

	view = NewHTMLView()
	items := items.New(view, view, db)
	
	var sync func()
	sync = func() {
		db.Sync([]idb.Model{items})
		time.AfterFunc(60 * time.Second, sync)
	}
	sync()

	document := js.Global.Get("document")

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
