package index

import (
	"appengine"
	"appengine/user"
	"encoding/json"
	"errors"
	"github.com/hajimehoshi/kakeibo/models"
	"html/template"
	"io/ioutil"
	"net/http"
)

// TODO: Use memcache?

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

func parseRequest(r *http.Request) (
	req *models.SyncRequest,
	reqItems []*models.ItemData,
	err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	req = &models.SyncRequest{}
	if err = json.Unmarshal(body, &req); err != nil {
		return
	}
	reqItems = []*models.ItemData{}
	for _, v := range req.Values {
		v, ok := v.(*models.ItemData)
		if !ok {
			err = errors.New("invalid data")
			return
		}
		reqItems = append(reqItems, v)
	}
	return
}

func handleSync(w http.ResponseWriter, r *http.Request) {
	req, reqItems, err := parseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c := appengine.NewContext(r)
	u := user.Current(c)

	d := NewItemDatastore(c, u.ID)
	now, err := d.Put(req.LastUpdated, reqItems)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resItems, err := d.Get(req.LastUpdated)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	values := make([]interface{}, len(resItems))
	for i, v := range resItems {
		values[i] = v
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
