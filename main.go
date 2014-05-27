// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/idb"
	"github.com/hajimehoshi/kakeibo/items"
	"github.com/hajimehoshi/kakeibo/view"
	"time"
)

func printError(err error) {
	js.Global.Get("console").Call("error", err.Error())
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

func ready() {
	db := idb.New(dbName, printError)

	v := view.NewHTMLView()
	items := items.New(v, db)
	
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
