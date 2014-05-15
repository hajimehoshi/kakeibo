// +build js

package main

import (
	"encoding/json"
	"fmt"
	"github.com/gopherjs/gopherjs/js"
)

type IDB struct {
	db       js.Object
	observer IDBObserver
}

type IDBObserver interface {
	OnReady(d *IDB)
	//OnValueUpdated()
}

func onError(e js.Object) {
	err := e.Get("target").Get("error")
	name := err.Get("name").Str()
	msg := err.Get("message").Str()
	print(fmt.Sprintf("%s: %s", name, msg))
}

func NewIDB(name string, observer IDBObserver) *IDB {
	idb := &IDB{
		db:       nil,  
		observer: observer,
	}

	const version = 1
	req := js.Global.Get("indexedDB").Call("open", name, version)
	req.Set("onupgradeneeded", func(e js.Object) {
		db := e.Get("target").Get("result")
		db.Call("createObjectStore", "items", map[string]interface{}{
			"keyPath": "id",
			"autoIncrement": false,
		})
		// FIXME: Create indexes
	})
	req.Set("onsuccess", func(e js.Object) {
		idb.db = e.Get("target").Get("result")
		idb.db.Set("onerror", onError)
		if idb.observer != nil {
			idb.observer.OnReady(idb)
		}
	})
	req.Set("onerror", onError)

	return idb
}

func (i *IDB) IsReady() bool {
	return i.db != nil
}

func (i *IDB) Save(value interface{}) {
	db := i.db
	t := db.Call("transaction", "items", "readwrite")
	s := t.Call("objectStore", "items")

	// TODO: Use JSON.stringify here?
	valStr, err := json.Marshal(value)
	if err != nil {
		print(err.Error())
		return
	}
	j := js.Global.Get("JSON").Call("parse", string(valStr))
	req := s.Call("put", j)
	req.Set("onsuccess", func() {
		print("OK!")
		// FIXME: call callback
	})
	req.Set("onerror", onError)
}
