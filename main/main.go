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
	v.SetItems(items)

	db.Init([]idb.Model{items})

	var sync func()
	sync = func() {
		db.SyncIfNeeded([]idb.Model{items})
		time.AfterFunc(10 * time.Second, sync) 
	}
	sync()

	/*worker := js.Global.Get("SharedWorker").New("/static/scripts/shared.js")
	worker.Get("port").Set("onmessage", func(e js.Object) {
		print(e.Get("data").Str())
	})
	worker.Get("port").Call("postMessage", "foooo")*/

	document := js.Global.Get("document")
	debugLink := document.Call("getElementById", "debug_link")
	debugLink.Set("onclick", toggleDebugOverlay)
	debugOverlay := document.Call("getElementById", "debug_overlay")
	debugOverlay.Set("onclick", toggleDebugOverlay)

	js.Global.Get("window").Set("onhashchange", v.OnHashChange)
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
