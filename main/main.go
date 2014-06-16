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

// FIXME: Use appengine.IsDevAppServer() instead
func isDevelopment() bool {
	hostname := js.Global.Get("location").Get("hostname").Str()
	return hostname == "localhost" || hostname == "127.0.0.1"
}

func ready() {
	// TODO: Don't use IndexedDB (if needed).
	// Or, create shared worker.
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

	document := js.Global.Get("document")

	if isDevelopment() {
		ds := document.Call("querySelectorAll", "span.development")
		for i := 0; i < ds.Length(); i++ {
			d := ds.Index(i)
			d.Get("style").Set("display", "inline")
		}

		m := document.Call("getElementById", "mode")
		m.Set("textContent", "(Development Mode)")

		// TODO: This grid mode can be public even on the production
		// mode.
		debugLink := document.Call("getElementById", "debug_link")
		debugLink.Set("onclick", toggleDebugOverlay)
		debugOverlay := document.Call("getElementById", "debug_overlay")
		debugOverlay.Set("onclick", toggleDebugOverlay)
	}

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