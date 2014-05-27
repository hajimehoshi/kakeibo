package index

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hajimehoshi/kakeibo/models"
	"github.com/hajimehoshi/kakeibo/uuid"
	"html/template"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"
)

const (
	kindItems = "Items"
)

var (
	rootKeyStringID = reflect.TypeOf((*models.ItemData)(nil)).Elem().Name()
)

var tmpl *template.Template

func init() {
	http.HandleFunc("/sync", filterUsers(handleSync))
	http.HandleFunc("/", filterUsers(handleIndex))

	var err error
	tmpl, err = template.ParseFiles("templates/index.html")
	if err != nil {
		panic(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-type", "text/html; charset=utf-8")
	c := appengine.NewContext(r)
	u := user.Current(c)
	url, _ := user.LogoutURL(c, "/")
	tmpl.Execute(w, map[string]interface{}{
		"UserEmail": u.Email,
		"LogoutURL": url,
	})
}

// TODO: Split this into two works (PUT and GET)
// TODO: Use memcache?
func handleSync(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req := models.SyncRequest{}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	lastUpdated := req.LastUpdated
	reqItems := map[uuid.UUID]*models.ItemData{}
	for _, v := range req.Values {
		v, ok := v.(*models.ItemData)
		if !ok {
			http.Error(w, "invalid data", http.StatusBadRequest)
			return
		}
		reqItems[v.Meta.ID] = v
	}

	now := models.UnixTime(time.Now().Unix())
	if now < lastUpdated {
		http.Error(w, "last_updated is too new", http.StatusBadRequest)
		return
	}

	resItems := map[uuid.UUID]*models.ItemData{}

	if err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		serverItems := map[uuid.UUID]*models.ItemData{}
		rootKey := datastore.NewKey(c, kindItems, rootKeyStringID, 0, nil)
		q := datastore.NewQuery(kindItems)
		q = q.Ancestor(rootKey)
		q = q.Filter("Meta.LastUpdated >=", lastUpdated)
		q = q.Filter("Meta.UserID =", u.ID)
		t := q.Run(c)
		for {
			var d models.ItemData
			_, err := t.Next(&d)
			if err == datastore.Done {
				break
			}
			if err != nil {
				return err
			}
			serverItems[d.Meta.ID] = &d
		}

		serverNewItems := []*models.ItemData{}
		for id, d := range serverItems {
			resItems[id] = d
		}
		for id, d := range reqItems {
			if _, ok := serverItems[id]; ok {
				continue
			}
			// The requested data is new.
			strID := id.String()
			key := datastore.NewKey(c, kindItems, strID, 0, rootKey)
			var d2 models.ItemData
			err := datastore.Get(c, key, &d2)
			if err == nil {
				if d2.Meta.UserID != u.ID {
					return errors.New(
						fmt.Sprintf("invalid UUID: %s", strID))
				}
			} else if err != datastore.ErrNoSuchEntity {
				return err
			}
			d.Meta.LastUpdated = now
			d.Meta.UserID = u.ID
			serverNewItems = append(serverNewItems, d)
			resItems[id] = d
		}

		keys := []*datastore.Key{}
		for _, d := range serverNewItems {
			strID := d.Meta.ID.String()
			keys = append(keys, datastore.NewKey(c, kindItems, strID, 0, rootKey))
		}
		if _, err := datastore.PutMulti(c, keys, serverNewItems); err != nil {
			return err
		}

		return nil
	}, nil); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	values := []interface{}{}
	for _, v := range resItems {
		values = append(values, v)
	}
	res := &models.SyncResponse{
		Type: req.Type,
		LastUpdated: now,
		Values: values,
	}
	resBytes, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(resBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

