package hello

import (
	"appengine"
	"appengine/user"
	"fmt"
	"github.com/hajimehoshi/kakeibo/date"
	"net/http"
)

func init() {
	http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	fmt.Fprintf(w, "Hello, %s! Today: %s", u.ID, date.Today())
	url, _ := user.LogoutURL(c, "/")
	fmt.Fprintf(w, "<br /><a href=\"%s\">Logout</a>", url)
}
