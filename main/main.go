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

func deleteDBIfUserChanged(name string) chan struct{} {
	ch := make(chan struct{})
	ls := js.Global.Get("localStorage")
	last := ls.Call("getItem", "last_user_email").Str()
	current := js.Global.Call("userEmail").Str()
	if last == current {
		close(ch)
		return ch
	}
	req := js.Global.Get("indexedDB").Call("deleteDatabase", name)
	req.Set("onsuccess", func() {
		close(ch)
	})
	ls.Call("setItem", "last_user_email", current)
	return ch
}

const dbName = "kakeibo"

func main() {
	<-deleteDBIfUserChanged(dbName)

	// TODO: Don't use IndexedDB (if needed).
	// Or, create shared worker.
	db := idb.New(dbName, printError)

	v := view.NewHTMLView(printError)
	items := items.New(v, db)
	v.SetItems(items)

	<-db.Init([]idb.Model{items})

	document := js.Global.Get("document")

	if js.Global.Call("isDevelopmentMode").Bool() {
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

	for {
		db.SyncIfNeeded([]idb.Model{items})
		time.Sleep(10 * time.Second)
	}
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
