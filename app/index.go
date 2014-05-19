package index

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hajimehoshi/kakeibo/models"
	"github.com/hajimehoshi/kakeibo/uuid"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	kindItems = "Items"
)

var tmpl *template.Template

func init() {
	http.HandleFunc("/sync", handleSync)
	http.HandleFunc("/", handleIndex)

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

func lines(r io.Reader) ([]string, error) {
	lines := []string{}
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadString('\n')
		if err == nil || err == io.EOF {
			line = strings.Trim(line, " \t\v\r\n")
			lines = append(lines, line)
			if err == io.EOF {
				break
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return lines, nil
}

func handleSync(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)

	// request:
	// {type: "items", last_updated: 12345}
	// {JSON}
	// {JSON}
	//
	// response:
	// {type: "items", last_updated: 67890}
	// {JSON + last_updated}
	// {JSON + last_updated}

	ls, err := lines(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(ls) == 0 {
		http.Error(w, "empty request", http.StatusBadRequest)
		return
	}

	reqItems := map[uuid.UUID]*models.ItemData{}
	for _, l := range ls[1:] {
		var d models.ItemData
		if err := json.Unmarshal([]byte(l), &d); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if !d.IsValid() {
			continue
		}
		reqItems[d.Meta.ID] = &d
	}

	meta := map[string]string{}
	if err := json.Unmarshal([]byte(ls[0]), &meta); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	lastUpdated := models.UnixTime(0)
	if err := lastUpdated.UnmarshalText([]byte(meta["last_updated"])); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	now := models.UnixTime(time.Now().Unix())
	if now < lastUpdated {
		http.Error(w, "last_updated is too new", http.StatusBadRequest)
		return
	}

	resItems := map[uuid.UUID]*models.ItemData{}

	if err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		const rootKeyStringID = "items"

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
				return errors.New("invalid UUID")
			}
			if err != datastore.ErrNoSuchEntity {
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

	j, err := json.Marshal(map[string]interface{}{
		"type":         "items",
		"last_updated": now,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s\n", j)
	for _, i := range resItems {
		j, err := json.Marshal(i)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%s\n", j)
	}
}
