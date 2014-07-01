// +build js

package idb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/models"
	"reflect"
	"time"
)

// TODO: Use SharedWorker.

const (
	lastUpdatedIndex = "LastUpdated"
)

type Model interface {
	Type() reflect.Type
	OnLoaded(vals []interface{})
}

type IDB struct {
	name         string
	onErrorFunc  func(error)
	db           js.Object
	lastUpdated  time.Time
	syncNeeded   bool
}

func toError(e js.Object) error {
	jsErr := e.Get("error")
	name := jsErr.Get("name").Str()
	msg := jsErr.Get("message").Str()
	return errors.New(fmt.Sprintf("idb: %s: %s", name, msg))
}

func DeleteDBIfUserChanged(name string) error {
	ch := make(chan error)

	ls := js.Global.Get("localStorage")
	last := ls.Call("getItem", "last_user_email").Str()
	current := js.Global.Call("userEmail").Str()
	if last == current {
		close(ch)
		return nil
	}

	req := js.Global.Get("indexedDB").Call("deleteDatabase", name)
	req.Set("onsuccess", func(e js.Object) {
		close(ch)
	})
	req.Set("onerror", func(e js.Object) {
		go func() {
			ch <- toError(e.Get("target"))
			close(ch)
		}()
	})
	ls.Call("setItem", "last_user_email", current)

	if err := <-ch; err != nil {
		return err
	}
	return nil
}

func New(name string, onErrorFunc func(error)) *IDB {
	return &IDB{
		name:        name,
		onErrorFunc: onErrorFunc,
		syncNeeded:  true,
	}
}

func (i *IDB) onError(err error) {
	if i.onErrorFunc == nil {
		return
	}
	i.onErrorFunc(err)
}

func (i *IDB) idxOnError(e js.Object) {
	i.onError(toError(e.Get("target")))
}

func (i *IDB) Save(value interface{}) error {
	i.syncNeeded = true
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
	req.Set("onerror", i.idxOnError)
	return nil
}

func jsonStringify(v interface{}) string {
	return js.Global.Get("JSON").Call("stringify", v).Str()
}

func (i *IDB) loadAll(m Model) error {
	db := i.db
	t := m.Type()
	tr := db.Call("transaction", t.Name(), "readonly")
	s := tr.Call("objectStore", t.Name())
	req := s.Call("openCursor")
	values := []interface{}{}
	req.Set("onsuccess", func(e js.Object) {
		cursor := e.Get("target").Get("result")
		if cursor.IsNull() {
			m.OnLoaded(values)
			return
		}
		value := cursor.Get("value")
		if value.Get("Meta").Get("IsDeleted").Bool() {
			cursor.Call("continue")
			return
		}
		v := reflect.New(t).Interface()
		j := jsonStringify(value)
		if err := json.Unmarshal([]byte(j), &v); err != nil {
			i.onError(err)
			return
		}
		values = append(values, v)
		cursor.Call("continue")
	})
	req.Set("onerror", i.idxOnError)

	return nil
}

func (i *IDB) SyncIfNeeded(models []Model) error {
	if !i.syncNeeded {
		return nil
	}
	for _, m := range models {
		if err := i.initLastUpdated(m); err != nil {
			return err
		}
		i.getUnsyncedItems(m, i.sync)
	}
	i.syncNeeded = false
	return nil
}

func (i *IDB) Init(models []Model) error {
	ch := make(chan error)

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
		// TODO: Setting is necessary?
		/*i.db.Set("onerror", func(e js.Object) {
			ch <- toError(e.Get("target"))
		})*/

		for _, m := range models {
			i.loadAll(m)
		}
		close(ch)
	})
	req.Set("onerror", func(e js.Object) {
		go func() {
			ch <- toError(e.Get("target"))
			close(ch)
		}()
	})

	if err := <-ch; err != nil {
		return err
	}
	return nil
}

func (i *IDB) initLastUpdated(m Model) error {
	if !i.lastUpdated.IsZero() {
		return nil
	}

	ch := make(chan error)

	db := i.db
	t := m.Type()
	tr := db.Call("transaction", t.Name(), "readonly")
	s := tr.Call("objectStore", t.Name())
	idx := s.Call("index", lastUpdatedIndex)
	req := idx.Call("openCursor", nil, "prev")
	req.Set("onsuccess", func(e js.Object) {
		maxLastUpdated := time.Time{}

		cursor := e.Get("target").Get("result")
		if !cursor.IsNull() {
			value := cursor.Get("value")
			l := value.Get("Meta").Get("LastUpdated").Str()
			if err := maxLastUpdated.UnmarshalText([]byte(l));
			err != nil {
				go func() {
					ch <- err
					close(ch)
				}()
				return
			}
			if i.lastUpdated.Before(maxLastUpdated) {
				i.lastUpdated = maxLastUpdated
			}
		}
	})
	req.Set("onerror", func(e js.Object) {
		go func() {
			ch <- toError(e.Get("target"))
			close(ch)
		}()
	})

	if err := <-ch; err != nil {
		return err
	}
	return nil
}

func (i *IDB) getUnsyncedItems(m Model, f func(Model, []interface{})) {
	// A record whose LastUpdated is zero time means a record which is not
	// synced.
	zerot, _ := time.Time{}.MarshalText()

	db := i.db
	t := m.Type()
	tr := db.Call("transaction", t.Name(), "readonly")
	s := tr.Call("objectStore", t.Name())
	idx := s.Call("index", lastUpdatedIndex)
	req := idx.Call("openCursor", string(zerot))
	values := []interface{}{}
	req.Set("onsuccess", func(e js.Object) {
		cursor := e.Get("target").Get("result")
		if cursor.IsNull() {
			f(m, values)
			return
		}
		j := cursor.Get("value")
		jStr := jsonStringify(j)
		value := reflect.New(t).Interface()
		if err := json.Unmarshal([]byte(jStr), &value); err != nil {
			i.onError(err)
			cursor.Call("continue")
			return
		}
		values = append(values, value)
		cursor.Call("continue")
		return
	})
	req.Set("onerror", i.idxOnError)
}

func (i *IDB) sync(m Model, values []interface{}) {
	req := js.Global.Get("XMLHttpRequest").New()
	req.Call("open", "POST", "/sync", true)
	req.Set("onload", func(e js.Object) {
		xhr := e.Get("target")
		if xhr.Get("status").Int() != 200 {
			return
		}
		text := xhr.Get("responseText").Str()
		res := models.SyncResponse{}
		if err := json.Unmarshal([]byte(text), &res); err != nil {
			i.onError(err)
			return
		}
		vals := []interface{}{}
		for _, v := range res.Values {
			v, ok := v.(*models.ItemData)
			if !ok {
				err := errors.New("idb: invalid response")
				i.onError(err)
				return
			}
			if err := i.put(v); err != nil {
				i.onError(err)
				return
			}
			vals = append(vals, v)
		}
		i.lastUpdated = res.LastUpdated
		m.OnLoaded(vals)
	})
	req.Set("onerror", func(e js.Object) {
		i.onError(toError(e))
	})
	request := models.SyncRequest{
		Type:        m.Type().Name(),
		LastUpdated: i.lastUpdated,
		Values:      values,
	}
	str, _ := json.Marshal(request)
	req.Call("send", str)
}
