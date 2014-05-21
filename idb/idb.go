// +build js

package idb

import (
	"encoding/json"
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/models"
	"reflect"
	"sync"
)

const (
	lastUpdatedIndex = "LastUpdated"
)

type Model interface {
	Type() reflect.Type
	OnLoaded(vals []interface{})
	OnInitialLoaded(vals []interface{})
}

type IDB struct {
	name         string
	initializing sync.Once
	db           js.Object
	queue        []func()
}

func onError(e js.Object) {
	err := e.Get("target").Get("error")
	name := err.Get("name").Str()
	msg := err.Get("message").Str()
	print(fmt.Sprintf("%s: %s", name, msg))
}

func New(name string) *IDB {
	idb := &IDB{
		name:  name,
		db:    nil,
		queue: []func(){},
	}
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

	return i.put(value)
}

func (i *IDB) put(v interface{}) error {
	json, err := json.Marshal(v)
	if err != nil {
		return err
	}
	j := js.Global.Get("JSON").Call("parse", string(json))
	db := i.db
	t := reflect.TypeOf(v).Elem()
	tr := db.Call("transaction", t.Name(), "readwrite")
	s := tr.Call("objectStore", t.Name())
	req := s.Call("put", j)
	req.Set("onerror", onError)
	return nil
}

func jsonStringify(v interface{}) string {
	return js.Global.Get("JSON").Call("stringify", v).Str()
}

func (i *IDB) loadAll(m Model) error {
	if !i.isReady() {
		i.queue = append(i.queue, func() {
			i.loadAll(m)
		})
		return nil
	}

	db := i.db
	t := m.Type()
	tr := db.Call("transaction", t.Name(), "readonly")
	s := tr.Call("objectStore", t.Name())
	req := s.Call("openCursor")
	values := []interface{}{}
	req.Set("onsuccess", func(e js.Object) {
		cursor := e.Get("target").Get("result")
		if cursor.IsNull() {
			m.OnInitialLoaded(values)
			return
		}
		value := cursor.Get("value")
		if value.Get("Meta").Get("IsDeleted").Bool() {
			cursor.Call("continue")
			return
		}
		v := reflect.New(t).Interface()
		j := jsonStringify(value)
		if err := json.Unmarshal([]byte(j), v); err != nil {
			// TODO: fix this
			print(err.Error())
			return
		}
		values = append(values, v)
		cursor.Call("continue")
	})
	req.Set("onerror", onError)

	return nil
}

func (i *IDB) Sync(models []Model) {
	if !i.isReady() {
		i.initializing.Do(func() {
			i.init(models)
		})
		i.queue = append(i.queue, func() {
			i.Sync(models)
		})
		return
	}
	for _, m := range models {
		i.sync(m)
	}
}

func (i *IDB) init(models []Model) {
	const version = 1
	req := js.Global.Get("indexedDB").Call("open", i.name, version)
	req.Set("onupgradeneeded", func(e js.Object) {
		db := e.Get("target").Get("result")
		for _, m := range models {
			store := db.Call(
				"createObjectStore",
				m.Type().Name(),
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
		i.db = e.Get("target").Get("result")
		i.db.Set("onerror", onError)

		for _, m := range models {
			i.loadAll(m)
		}

		for _, f := range i.queue {
			f()
		}
		i.queue = []func(){}
	})
	req.Set("onerror", onError)
}

func (i *IDB) sync(m Model) {
	// FIXME: Save this as a member variable. Don't use the same value
	// repeatedly.
	maxLastUpdated := models.UnixTime(0)

	db := i.db
	t := m.Type()
	tr := db.Call("transaction", t.Name(), "readonly")
	s := tr.Call("objectStore", t.Name())
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
				i.sync2(m, maxLastUpdated, values)
				return
			}
			j := cursor.Get("value")
			jStr := jsonStringify(j)
			value := reflect.New(t).Interface()
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
	m Model,
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
		vals := []interface{}{}
		for _, v := range res.Values {
			v, ok := v.(*models.ItemData)
			if !ok {
				// TODO: Fix this
				print("invalid response")
				return
			}
			if err := i.put(v); err != nil {
				// TODO: Fix this
				print(err.Error())
				return
			}
			vals = append(vals, v)
		}
		m.OnLoaded(vals)
	})
	req.Set("onerror", func(e js.Object) {
		// FIXME: implement this
	})
	request := models.SyncRequest{
		Type:        m.Type().Name(),
		LastUpdated: maxLastUpdated,
		Values:      values,
	}
	str, _ := json.Marshal(request)
	req.Call("send", str)
}
