// +build js

package idb

import (
	"encoding/json"
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"reflect"
)

type SchemaNotFoundError struct {
	Value interface{}
}

func (e *SchemaNotFoundError) Error() string {
	return "schema not found"
}

type InvalidValueError struct {
	Inner error
	Value interface{}
}

func (e *InvalidValueError) Error() string {
	return "invalid value"
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
	observer  IDBObserver
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

func New(name string, schemaSet SchemaSet, observer IDBObserver) *IDB {
	idb := &IDB{
		db:        nil,  
		schemaSet: schemaSet,
		observer:  observer,
	}

	const version = 1
	req := js.Global.Get("indexedDB").Call("open", name, version)
	req.Set("onupgradeneeded", func(e js.Object) {
		db := e.Get("target").Get("result")
		for _, schema := range idb.schemaSet {
			db.Call(
				"createObjectStore",
				schema.Name,
				map[string]interface{}{
					"keyPath": "meta.id",
					"autoIncrement": false,
				})
			// FIXME: Create index for last_updated
		}
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

func (i *IDB) Save(value interface{}) error {
	schema := i.schemaSet.GetFor(value)
	if schema == nil {
		return &SchemaNotFoundError{value}
	}

	db := i.db
	t := db.Call("transaction", schema.Name, "readwrite")
	s := t.Call("objectStore", schema.Name)

	// TODO: Use JSON.stringify here?
	valStr, err := json.Marshal(value)
	print(string(valStr))
	if err != nil {
		return &InvalidValueError{err, value}
	}
	j := js.Global.Get("JSON").Call("parse", string(valStr))
	req := s.Call("put", j)
	req.Set("onsuccess", func() {
		print("OK!")
		// FIXME: call callback
	})
	req.Set("onerror", onError)

	return nil
}

// TODO: Syncing with the server
