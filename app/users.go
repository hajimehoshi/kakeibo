package index

import (
	"appengine"
	"appengine/user"
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var permittedUserEmails = map[string]struct{}{}

var forbiddenTmpl *template.Template
var forbiddenTmplStr = `
<!DOCTYPE html>
<p>Forbidden (<a href="{{.LogoutURL}}">Logout</a>)</p>
`[1:]

func init() {
	const filename = "users.txt"
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	content, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	lines := bytes.Split(content, []byte("\n"))
	for _, l := range lines {
		l := strings.Trim(string(l), " \r\n\t\v")
		if len(l) == 0 || l[0] == '#' {
			continue
		}
		permittedUserEmails[l] = struct{}{}
	}

	forbiddenTmpl, err = template.New("forbidden").Parse(forbiddenTmplStr)
	if err != nil {
		panic(err)
	}
}

func filterUsers(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		u := user.Current(c)
		if u == nil {
			w.Header().Set(
				"Content-type",
				"text/plain; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "'login: required' is needed at app.yaml.\n")
			return
		}
		if _, ok := permittedUserEmails[u.Email]; !ok {
			w.Header().Set(
				"Content-type",
				"text/html; charset=utf-8")
			url, _ := user.LogoutURL(c, "/")
			w.WriteHeader(http.StatusForbidden)
			forbiddenTmpl.Execute(w, map[string]interface{}{
				"LogoutURL": url,
			})
			return
		}
		f(w, r)
	}
}
