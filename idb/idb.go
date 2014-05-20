// +build js

package idb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/models"
	"github.com/hajimehoshi/kakeibo/uuid"
	"reflect"
)

const (
	lastUpdatedIndex = "last_updated"
)

type Schema struct {
	Type reflect.Type
	Name string
}

type SchemaSet struct {
	set map[reflect.Type]*Schema
}

func NewSchemaSet() *SchemaSet {
	return &SchemaSet{
		set: map[reflect.Type]*Schema{},
	}
}

func (s *SchemaSet) Add(schema *Schema) {
	s.set[schema.Type] = schema
}

func (s *SchemaSet) GetFor(value interface{}) *Schema {
	t := reflect.TypeOf(value)
	if schema, ok := s.set[t]; ok {
		return schema
	}
	for t2, schema := range s.set {
		if reflect.PtrTo(t2) == t {
			return schema
		}
	}
	return nil
}

type IDB struct {
	db        js.Object
	schemaSet *SchemaSet
	queue     []func()
}

func onError(e js.Object) {
	err := e.Get("target").Get("error")
	name := err.Get("name").Str()
	msg := err.Get("message").Str()
	print(fmt.Sprintf("%s: %s", name, msg))
}

func New(name string, schemaSet *SchemaSet) *IDB {
	idb := &IDB{
		db:        nil,
		schemaSet: schemaSet,
		queue:     []func(){},
	}

	const version = 1
	req := js.Global.Get("indexedDB").Call("open", name, version)
	req.Set("onupgradeneeded", func(e js.Object) {
		db := e.Get("target").Get("result")
		for _, schema := range idb.schemaSet.set {
			store := db.Call(
				"createObjectStore",
				schema.Name,
				map[string]interface{}{
					"keyPath":       "Meta.ID",
					"autoIncrement": false,
				})
			store.Call(
				"createIndex",
				lastUpdatedIndex,
				"Meta.LastUpdated",
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
		return errors.New("IDB.Save: schema not found")
	}

	if err := i.put(schema, value); err != nil {
		return err
	}
	return nil
}

func (i *IDB) put(schema *Schema, v interface{}) error {
	json, err := json.Marshal(v)
	if err != nil {
		return err
	}
	j := js.Global.Get("JSON").Call("parse", string(json))
	db := i.db
	t := db.Call("transaction", schema.Name, "readwrite")
	s := t.Call("objectStore", schema.Name)
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
		if value.Get("Meta").Get("IsDeleted").Bool() {
			cursor.Call("continue")
			return
		}
		values = append(values, jsonStringify(value))
		cursor.Call("continue")
	})
	req.Set("onerror", onError)

	return nil
}

func (i *IDB) Sync() {
	if !i.isReady() {
		i.queue = append(i.queue, func() {
			i.Sync()
		})
		return
	}
	for _, s := range i.schemaSet.set {
		i.sync(s)
	}
}

func (i *IDB) sync(schema *Schema) {
	// FIXME: Save this as a member variable. Don't use the same value
	// repeatedly.
	maxLastUpdated := models.UnixTime(0)

	db := i.db
	t := db.Call("transaction", schema.Name, "readonly")
	s := t.Call("objectStore", schema.Name)
	idx := s.Call("index", lastUpdatedIndex)
	// FIXME: Use the openCursor iff |maxLastUpdated| == 0.
	req := idx.Call("openCursor", nil, "prev")
	req.Set("onsuccess", func(e js.Object) {
		cursor := e.Get("target").Get("result")
		if !cursor.IsNull() {
			value := cursor.Get("value")
			l := value.Get("Meta").Get("LastUpdated").Str()
			if err := maxLastUpdated.UnmarshalText([]byte(l)); err != nil {
				// TODO: Fix this
				print(err.Error())
				return
			}
		}
		req := idx.Call("openCursor", models.UnixTime(0).String())
		values := []interface{}{}
		req.Set("onsuccess", func(e js.Object) {
			cursor := e.Get("target").Get("result")
			if cursor.IsNull() {
				i.sync2(schema, maxLastUpdated, values)
				return
			}
			j := cursor.Get("value")
			jStr := jsonStringify(j)
			value := reflect.New(schema.Type).Interface()
			if err := json.Unmarshal([]byte(jStr), value); err != nil {
				// TODO: Fix this
				print(err.Error())
				cursor.Call("continue")
				return
			}
			values = append(values, value)
			cursor.Call("continue")
			return
		})
		req.Set("onerror", onError)
	})
	req.Set("onerror", onError)
}

func (i *IDB) sync2(
	schema *Schema,
	maxLastUpdated models.UnixTime,
	values []interface{}) {
	req := js.Global.Get("XMLHttpRequest").New()
	req.Call("open", "POST", "/sync", true)
	req.Set("onload", func(e js.Object) {
		xhr := e.Get("target")
		if xhr.Get("status").Int() != 200 {
			// TODO: Fix this
			print(xhr.Get("responseText").Str())
			return
		}
		text := xhr.Get("responseText").Str()
		res := models.SyncResponse{}
		if err := json.Unmarshal([]byte(text), &res); err != nil {
			// TODO: Fix this
			print(err.Error())
			return
		}
		for _, v := range res.Values {
			v, ok := v.(*models.ItemData)
			if !ok {
				// TODO: Fix this
				print("invalid response")
				return
			}
			if err := i.put(schema, v); err != nil {
				// TODO: Fix this
				print(err.Error())
				return
			}
		}
	})
	req.Set("onerror", func(e js.Object) {
		// FIXME: implement this
	})
	request := models.SyncRequest{
		Type:        schema.Name,
		LastUpdated: maxLastUpdated,
		Values:      values,
	}
	str, _ := json.Marshal(request)
	req.Call("send", str)
}
