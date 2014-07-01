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

const dbName = "kakeibo"

func main() {
	if err := idb.DeleteDBIfUserChanged(dbName); err != nil {
		printError(err)
		return
	}

	// TODO: Don't use IndexedDB (if needed).
	// Or, create shared worker.
	db := idb.New(dbName)

	v := view.NewHTMLView(printError)
	items := items.New(v, db)
	v.SetItems(items)

	if err := db.Init([]idb.Model{items}); err != nil {
		printError(err)
		return
	}

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
		err := db.SyncIfNeeded([]idb.Model{items})
		if err != nil {
			printError(err)
			return
		}
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
