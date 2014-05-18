package index

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/hajimehoshi/kakeibo/models"
	"github.com/hajimehoshi/kakeibo/uuid"
	"html/template"
	"io"
	"net/http"
	"strings"
)

const salt = "kakeibo"
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

	// request:
	// {type: "items", last_updated: 12345}
	// {JSON}
	// {JSON}
	//
	// response:
	// {type: "items", last_updated: 67890}
	// {JSON + last_updated}
	// {JSON + last_updated}
	// 
	// The client accepts the response and update all data. Then, the items
	// 'last_updated = 0' don't exist.
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

	meta := map[string]interface{}{}
	if err := json.Unmarshal([]byte(ls[0]), &meta); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	lastUpdated := models.UnixTime(meta["last_updated"].(float64))
	
	if err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		serverItems := map[uuid.UUID]*models.ItemData{}
		// FIXME: Check user
		key := datastore.NewKey(c, "Items", "items", 0, nil)
		q := datastore.NewQuery("Items")
		q = q.Ancestor(key).Filter("Meta.LastUpdated >=", lastUpdated)
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

		for id, d := range reqItems {
			if d2, ok := serverItems[id]; ok {
				if lastUpdated <= d2.Meta.LastUpdated {
					print(d)
					print(d2)
				}
			}
		}

		return nil
	}, nil); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	//if vals

	//d := []*models.ItemData
	// Check the data
	// Add user
	// Update db
	// Get data whose user is the current user
	// Return them
	fmt.Fprintf(w, "%+v\n", reqItems)
}
