// +build js

package idb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/uuid"
	"reflect"
	"strings"
)

const (
	lastUpdatedIndex = "last_updated"
)

type Schema struct {
	Name string
}

type SchemaSet map[reflect.Type]*Schema

func NewSchemaSet() SchemaSet {
	return SchemaSet(map[reflect.Type]*Schema{})
}

func (s SchemaSet) Add(t reflect.Type, schema *Schema) {
	s[t] = schema
}

func (s SchemaSet) GetFor(value interface{}) *Schema {
	return s[reflect.TypeOf(value)]
}

type IDB struct {
	db        js.Object
	schemaSet SchemaSet
	queue     []func()
}

func onError(e js.Object) {
	err := e.Get("target").Get("error")
	name := err.Get("name").Str()
	msg := err.Get("message").Str()
	print(fmt.Sprintf("%s: %s", name, msg))
}

func New(name string, schemaSet SchemaSet) *IDB {
	idb := &IDB{
		db:        nil,
		schemaSet: schemaSet,
		queue:     []func(){},
	}

	const version = 1
	req := js.Global.Get("indexedDB").Call("open", name, version)
	req.Set("onupgradeneeded", func(e js.Object) {
		db := e.Get("target").Get("result")
		for _, schema := range idb.schemaSet {
			store := db.Call(
				"createObjectStore",
				schema.Name,
				map[string]interface{}{
					"keyPath":       "meta.id",
					"autoIncrement": false,
				})
			store.Call(
				"createIndex",
				lastUpdatedIndex,
				"meta.last_updated",
				map[string]interface{}{
					"unique": false,
				})
			// TODO: create index for other columns
		}
	})
	req.Set("onsuccess", func(e js.Object) {
		idb.db = e.Get("target").Get("result")
		idb.db.Set("onerror", onError)
		for _, f := range idb.queue {
			f()
		}
		idb.queue = []func(){}
	})
	req.Set("onerror", onError)

	return idb
}

func (i *IDB) isReady() bool {
	return i.db != nil
}

func (i *IDB) Save(value interface{}) error {
	if !i.isReady() {
		i.queue = append(i.queue, func() {
			i.Save(value)
		})
		return nil
	}

	schema := i.schemaSet.GetFor(value)
	if schema == nil {
		return errors.New("idb: schema not found")
	}

	db := i.db
	t := db.Call("transaction", schema.Name, "readwrite")
	s := t.Call("objectStore", schema.Name)

	valStr, err := json.Marshal(value)
	if err != nil {
		return errors.New("idb: invalid value")
	}
	j := js.Global.Get("JSON").Call("parse", string(valStr))
	req := s.Call("put", j)
	req.Set("onerror", onError)

	return nil
}

func jsonStringify(v interface{}) string {
	return js.Global.Get("JSON").Call("stringify", v).Str()
}

func (i *IDB) Load(name string, id uuid.UUID, callback func(val string)) error {
	if !i.isReady() {
		i.queue = append(i.queue, func() {
			i.Load(name, id, callback)
		})
		return nil
	}

	db := i.db
	t := db.Call("transaction", name, "readonly")
	s := t.Call("objectStore", name)
	req := s.Call("get", id.String())
	req.Set("onsuccess", func(e js.Object) {
		result := e.Get("target").Get("result")
		callback(jsonStringify(result))
	})
	req.Set("onerror", onError)
	return nil
}

func (i *IDB) LoadAll(name string, callback func(val []string)) error {
	if !i.isReady() {
		i.queue = append(i.queue, func() {
			i.LoadAll(name, callback)
		})
		return nil
	}

	db := i.db
	t := db.Call("transaction", name, "readonly")
	s := t.Call("objectStore", name)
	req := s.Call("openCursor")
	values := []string{}
	req.Set("onsuccess", func(e js.Object) {
		cursor := e.Get("target").Get("result")
		if cursor.IsNull() {
			callback(values)
			return
		}
		value := cursor.Get("value")
		if value.Get("meta").Get("is_deleted").Bool() {
			cursor.Call("continue")
			return
		}
		values = append(values, jsonStringify(value))
		cursor.Call("continue")
	})
	req.Set("onerror", onError)

	return nil
}

func (i *IDB) Sync(name string) {
	if !i.isReady() {
		i.queue = append(i.queue, func() {
			i.Sync(name)
		})
		return
	}

	maxLastUpdated := float64(0)

	db := i.db
	t := db.Call("transaction", name, "readonly")
	s := t.Call("objectStore", name)
	idx := s.Call("index", lastUpdatedIndex)
	req := idx.Call("openCursor", nil, "prev")
	req.Set("onsuccess", func(e js.Object) {
		cursor := e.Get("target").Get("result")
		if !cursor.IsNull() {
			value := cursor.Get("value")
			maxLastUpdated =
				value.Get("meta").Get("last_updated").Float()
		}
		req := idx.Call("openCursor", 0)
		values := []js.Object{}
		req.Set("onsuccess", func(e js.Object) {
			cursor := e.Get("target").Get("result")
			if cursor.IsNull() {
				i.sync(name, maxLastUpdated, values)
				return
			}
			value := cursor.Get("value")
			values = append(values, value)
			cursor.Call("continue")
			return
		})
		req.Set("onerror", onError)
	})
	req.Set("onerror", onError)
}

func (i *IDB) sync(name string, maxLastUpdated float64, values []js.Object) {
	req := js.Global.Get("XMLHttpRequest").New()
	req.Call("open", "POST", "/sync", true)
	req.Set("onload", func(e js.Object) {
		xhr := e.Get("target")
		if xhr.Get("status").Int() != 200 {
			// FIXME: implement this
			print("error!")
			return
		}
		text := xhr.Get("responseText")
		print(text)
	})
	req.Set("onerror", func(e js.Object) {
		// FIXME: implement this
	})
	jsons := []string{
		jsonStringify(map[string]interface{}{
			"type":         name,
			"last_updated": maxLastUpdated,
		}),
	}
	for _, v := range values {
		jsons = append(jsons, jsonStringify(v))
	}
	req.Call("send", strings.Join(jsons, "\n"))
}
