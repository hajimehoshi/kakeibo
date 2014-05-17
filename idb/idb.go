// +build js

package idb

import (
	"encoding/json"
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/uuid"
	"reflect"
)

type SchemaNotFoundError struct {
	Value interface{}
}

func (e *SchemaNotFoundError) Error() string {
	return "idb: schema not found"
}

type InvalidValueError struct {
	Inner error
	Value interface{}
}

func (e *InvalidValueError) Error() string {
	return "idb: invalid value"
}

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
	buffer    []func()
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
		buffer:    []func(){},
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
				"last_updated",
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
		for _, f := range idb.buffer {
			f()
		}
		idb.buffer = []func(){}
	})
	req.Set("onerror", onError)

	return idb
}

func (i *IDB) isReady() bool {
	return i.db != nil
}

func (i *IDB) Save(value interface{}) error {
	if !i.isReady() {
		i.buffer = append(i.buffer, func() {
			i.Save(value)
		})
		return nil
	}

	schema := i.schemaSet.GetFor(value)
	if schema == nil {
		return &SchemaNotFoundError{value}
	}

	db := i.db
	t := db.Call("transaction", schema.Name, "readwrite")
	s := t.Call("objectStore", schema.Name)

	valStr, err := json.Marshal(value)
	if err != nil {
		return &InvalidValueError{err, value}
	}
	j := js.Global.Get("JSON").Call("parse", string(valStr))
	req := s.Call("put", j)
	req.Set("onerror", onError)

	return nil
}

func toJSONString(v js.Object) string {
	return js.Global.Get("JSON").Call("stringify", v).Str()
}

func (i *IDB) Load(name string, id uuid.UUID, callback func(val string)) error {
	if !i.isReady() {
		i.buffer = append(i.buffer, func() {
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
		callback(toJSONString(result))
	})
	req.Set("onerror", onError)
	return nil
}

func (i *IDB) LoadAll(name string, callback func(val []string)) error {
	if !i.isReady() {
		i.buffer = append(i.buffer, func() {
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
		if !cursor.IsNull() {
			value := cursor.Get("value")
			values = append(values, toJSONString(value))
			cursor.Call("continue")
			return
		}
		callback(values)
	})
	req.Set("onerror", onError)

	return nil
}

// TODO: Syncing with the server
