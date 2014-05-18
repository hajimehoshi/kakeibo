package index

import (
	"appengine"
	"appengine/user"
	"encoding/json"
	"fmt"
	//"github.com/hajimehoshi/kakeibo/date"
	"html/template"
	"io"
	"net/http"
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

func handleSync(w http.ResponseWriter, r *http.Request) {
	// request:
	// {type: "items", last_updated: "12345"}
	// {JSON}
	// {JSON}
	//
	// response:
	// {
	//   type: "items",
	//   last_updated: "12345",
	//   items: [
	//     {JSON + last_updated},
	//     {JSON + last_updated}
	//   ]
	// }
	// 
	// The client accepts the response and update all data. Then, the items
	// 'last_updated = 0' don't exist.
	dec := json.NewDecoder(r.Body)
	for {
		val := map[string]interface{}{}
		if err := dec.Decode(&val); err == io.EOF {
			break
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		fmt.Fprintf(w, "%+v\n", val)
	}
}
