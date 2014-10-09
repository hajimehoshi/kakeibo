package view

import (
	"errors"
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/items"
	"github.com/hajimehoshi/kakeibo/models"
	"github.com/hajimehoshi/kakeibo/uuid"
	"html"
	"reflect"
	"strconv"
	"strings"
)

type Items interface {
	UpdateDate(id uuid.UUID, date date.Date) error
	UpdateSubject(id uuid.UUID, subject string) error
	UpdateAmount(id uuid.UUID, amount int32) error
	Save(id uuid.UUID) error
	Destroy(id uuid.UUID) error
	UpdateMode(mode items.Mode, ym date.Date)
	DownloadCSV() error
}

// TODO: I18N

const (
	datasetAttrID  = "model-id"
	datasetAttrKey = "model-key"
)

func toDatasetProp(attr string) string {
	ts := strings.Split(attr, "-")
	if len(ts) == 0 {
		return ""
	}
	tokens := make([]string, len(ts))
	for i, t := range ts {
		if len(t) == 0 {
			continue
		}
		if i == 0 {
			tokens[i] = t
			continue
		}
		tokens[i] = strings.Title(t)
	}
	return strings.Join(tokens, "")
}

var (
	numberTypes = []reflect.Type{
		reflect.TypeOf((*int)(nil)).Elem(),
		reflect.TypeOf((*int8)(nil)).Elem(),
		reflect.TypeOf((*int16)(nil)).Elem(),
		reflect.TypeOf((*int32)(nil)).Elem(),
		reflect.TypeOf((*int64)(nil)).Elem(),
		reflect.TypeOf((*uint)(nil)).Elem(),
		reflect.TypeOf((*uint8)(nil)).Elem(),
		reflect.TypeOf((*uint16)(nil)).Elem(),
		reflect.TypeOf((*uint32)(nil)).Elem(),
		reflect.TypeOf((*uint64)(nil)).Elem(),
		reflect.TypeOf((*float32)(nil)).Elem(),
		reflect.TypeOf((*float64)(nil)).Elem(),
	}
)

func async(f func(e js.Object)) func(e js.Object) {
	return func(e js.Object) {
		e.Call("preventDefault")
		go func() {
			f(e)
		}()
	}
}

func isNumberType(t reflect.Type) bool {
	for _, nt := range numberTypes {
		if t.ConvertibleTo(nt) {
			return true
		}
	}
	return false
}

func getIDElement(e js.Object) js.Object {
	for {
		attr := e.Get("dataset").Get(toDatasetProp(datasetAttrID))
		if !attr.IsUndefined() {
			return e
		}
		e = e.Get("parentNode")
		if e.IsNull() || e.IsUndefined() {
			break
		}
	}
	return nil
}

func getIDFromElement(e js.Object) (uuid.UUID, error) {
	e2 := getIDElement(e)
	if e2 == nil {
		return *new(uuid.UUID), errors.New("view: element not found")
	}
	str := e2.Get("dataset").Get(toDatasetProp(datasetAttrID)).Str()
	id, err := uuid.ParseString(str)
	if err != nil {
		return *new(uuid.UUID), err
	}
	return id, nil
}

func printValueAt(e js.Object, name string, value string) {
	targets := []js.Object{}
	if e.Get("name").Str() == name {
		targets = append(targets, e)
	}
	query := fmt.Sprintf("*[name=\"%s\"]", html.EscapeString(name))
	es := e.Call("querySelectorAll", query)
	for i := 0; i < es.Length(); i++ {
		targets = append(targets, es.Index(i))
	}

	if e.Get("dataset").Get(toDatasetProp(datasetAttrKey)).Str() == name {
		targets = append(targets, e)
	}
	query = fmt.Sprintf(
		"*[data-%s=\"%s\"]",
		html.EscapeString(datasetAttrKey),
		html.EscapeString(name))
	es = e.Call("querySelectorAll", query)
	for i := 0; i < es.Length(); i++ {
		targets = append(targets, es.Index(i))
	}

	for _, e := range targets {
		if e.Call("hasAttribute", "value").Bool() {
			e.Set("value", value)
		} else {
			e.Set("textContent", value)
		}
	}
}

type HTMLView struct {
	items       Items
	onErrorFunc func(error)
}

func empty(e js.Object) {
	for e.Call("hasChildNodes").Bool() {
		e.Call("removeChild", e.Get("lastChild"))
	}
}

func NewHTMLView(onErrorFunc func(error)) *HTMLView {
	ch := make(chan js.Object)
	v := &HTMLView{
		onErrorFunc: onErrorFunc,
	}
	document := js.Global.Get("document")
	form := document.Call("getElementById", "form_item")
	form.Set("onsubmit", async(func(e js.Object) {
		go func() {
			ch <- e
		}()
	}))

	a := document.Call("getElementById", "link_export_as_csv")
	a.Set("onclick", async(v.onClickExportAsCSV))

	go func() {
		for e := range ch {
			switch e.Get("type").Str() {
			case "submit":
				v.onSubmit(e)
			}
		}
	}()

	return v
}

func (v *HTMLView) addEventListeners(items Items, form js.Object) {
	inputDate := form.Call("querySelector", "input[name=Date]")
	inputDate.Set("onchange", func(e js.Object) {
		id, err := getIDFromElement(e.Get("target"))
		if err != nil {
			v.onErrorFunc(err)
			return
		}
		dateStr := e.Get("target").Get("value").Str()
		d, err := date.ParseISO8601(dateStr)
		if err != nil {
			v.onErrorFunc(err)
			return
		}
		if err := items.UpdateDate(id, d); err != nil {
			v.onErrorFunc(err)
			return
		}
	})
	inputSubject := form.Call("querySelector", "input[name=Subject]")
	inputSubject.Set("onchange", func(e js.Object) {
		id, err := getIDFromElement(e.Get("target"))
		if err != nil {
			v.onErrorFunc(err)
			return
		}
		subject := e.Get("target").Get("value").Str()
		if err := items.UpdateSubject(id, subject); err != nil {
			v.onErrorFunc(err)
			return
		}
	})
	inputMoneyAmount := form.Call("querySelector", "input[name=Amount]")
	inputMoneyAmount.Set("onchange", func(e js.Object) {
		id, err := getIDFromElement(e.Get("target"))
		if err != nil {
			v.onErrorFunc(err)
			return
		}
		amount := int32(e.Get("target").Get("value").Int())
		if err := items.UpdateAmount(id, amount); err != nil {
			v.onErrorFunc(err)
			return
		}
	})
}

func (v *HTMLView) SetItems(items Items) {
	v.items = items
	document := js.Global.Get("document")
	form := document.Call("getElementById", "form_item")
	v.addEventListeners(items, form)
}

func removeSingleHash() {
	href := js.Global.Get("location").Get("href").Str()
	if 0 < len(href) && href[len(href)-1] == '#' {
		href = href[:len(href)-2]
		js.Global.Get("history").Call(
			"replaceState", "", "", href)
	}
}

func (v *HTMLView) OnHashChange(e js.Object) {
	hash := js.Global.Get("location").Get("hash").Str()
	// Remove the initial '#'
	if 1 <= len(hash) {
		hash = hash[1:]
	}
	switch hash {
	case "":
		removeSingleHash()
		v.updateMode(items.ModeTop, date.Date(0))
	default:
		ym, err := date.ParseISO8601(hash + "-01")
		if err != nil {
			v.onErrorFunc(err)
			return
		}
		v.updateMode(items.ModeYearMonth, ym)
	}
}

func (v *HTMLView) onSubmit(e js.Object) {
	form := e.Get("target")
	id, err := getIDFromElement(form)
	if err != nil {
		v.onErrorFunc(err)
		return
	}
	err = v.items.Save(id) //gopherjs:blocking
	if err != nil {
		v.onErrorFunc(err)
		return
	}
}

func (v *HTMLView) onClickExportAsCSV(e js.Object) {
	if err := v.items.DownloadCSV(); err != nil {
		v.onErrorFunc(err)
		return
	}
}

func (v *HTMLView) SetEditingItem(id uuid.UUID) {
	document := js.Global.Get("document")
	form := document.Call("getElementById", "form_item")
	form.Get("dataset").Set(toDatasetProp(datasetAttrID), id.String())
}

func (v *HTMLView) updateMode(mode items.Mode, ym date.Date) {
	v.items.UpdateMode(mode, ym)
}

func (v *HTMLView) PrintTitle(title string) {
	document := js.Global.Get("document")
	h1 := document.Call("querySelector", "body > main h1")
	h1.Set("textContent", title)
}

func (v *HTMLView) PrintItems(ids []uuid.UUID) {
	document := js.Global.Get("document")
	table := document.Call("getElementById", "table_items")
	tbody := table.Call("getElementsByTagName", "tbody").Index(0)
	empty(tbody)
	for _, id := range ids {
		v.addIDToItemTable(id)
	}
	display := "table"
	if len(ids) == 0 {
		display = "none"
	}
	table.Get("style").Set("display", display)
}

func (v *HTMLView) PrintItemsAndTotal(ids []uuid.UUID, total int) {
	v.PrintItems(ids)

	document := js.Global.Get("document")
	table := document.Call("getElementById", "table_items")
	tbody := table.Call("getElementsByTagName", "tbody").Index(0)

	tr := document.Call("createElement", "tr")

	td := document.Call("createElement", "td")
	td.Set("textContent", "")
	tr.Call("appendChild", td)

	td = document.Call("createElement", "td")
	td.Set("textContent", "(Total)")
	tr.Call("appendChild", td)

	td = document.Call("createElement", "td")
	td.Set("textContent", strconv.Itoa(int(total)))
	td.Get("classList").Call("add", "number")
	tr.Call("appendChild", td)

	td = document.Call("createElement", "td")
	td.Set("textContent", "")
	tr.Call("appendChild", td)

	tbody.Call("appendChild", tr)
}

func (v *HTMLView) PrintYearMonths(yms []date.Date) {
	document := js.Global.Get("document")
	ul := document.Call("getElementById", "year_months")
	empty(ul)
	for _, ym := range yms {
		a := document.Call("createElement", "a")
		date := fmt.Sprintf("%04d-%02d", ym.Year(), ym.Month())
		a.Set("textContent", date)
		a.Set("href", "#"+date)
		li := document.Call("createElement", "li")
		li.Call("appendChild", a)
		ul.Call("appendChild", li)
	}
}

func (v *HTMLView) PrintItem(data models.ItemData) {
	document := js.Global.Get("document")
	id := data.Meta.ID
	query := fmt.Sprintf(
		"*[data-%s=\"%s\"]",
		html.EscapeString(datasetAttrID),
		html.EscapeString(id.String()))
	elements := document.Call("querySelectorAll", query)
	for i := 0; i < elements.Length(); i++ {
		e := elements.Index(i)
		printValueAt(e, "Date", data.Date.String())
		printValueAt(e, "Subject", data.Subject)
		printValueAt(e, "Amount", strconv.Itoa(int(data.Amount)))
	}
}

func (v *HTMLView) addIDToItemTable(id uuid.UUID) {
	t := reflect.TypeOf((*models.ItemData)(nil)).Elem()

	document := js.Global.Get("document")
	table := document.Call("getElementById", "table_items")
	query := fmt.Sprintf("tr[data-%s=\"%s\"]", datasetAttrID, id.String())
	if 1 <= table.Call("querySelectorAll", query).Length() {
		return
	}
	tr := document.Call("createElement", "tr")
	tr.Get("dataset").Set(toDatasetProp(datasetAttrID), id.String())

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type == reflect.TypeOf((*models.Meta)(nil)).Elem() {
			continue
		}
		td := document.Call("createElement", "td")
		td.Get("dataset").Set(toDatasetProp(datasetAttrKey), f.Name)
		if isNumberType(f.Type) {
			td.Get("classList").Call("add", "number")
		}
		tr.Call("appendChild", td)
	}

	a := document.Call("createElement", "a")
	a.Set("textContent", "Delete")
	a.Call("setAttribute", "href", "")
	td := document.Call("createElement", "td")
	td.Call("appendChild", a)
	td.Get("classList").Call("add", "action")
	a.Set("onclick", async(v.onClickToDelete))
	tr.Call("appendChild", td)

	tbody := table.Call("getElementsByTagName", "tbody").Index(0)
	tbody.Call("appendChild", tr)
}

func (v *HTMLView) onClickToDelete(e js.Object) {
	id, err := getIDFromElement(e.Get("target"))
	if err != nil {
		v.onErrorFunc(err)
		return
	}
	// TODO: Show confirming alert if needed.
	go func() {
		err = v.items.Destroy(id) //gopherjs:blocking
		if err != nil {
			v.onErrorFunc(err)
			return
		}
	}()
}

func (v *HTMLView) Download(b []byte, filename string) {
	document := js.Global.Get("document")
	a := document.Call("createElement", "a")
	blob := js.Global.Get("Blob").New(
		[][]byte{b},
		map[string]string{
			"type": "application/octet-stream",
		},
	)
	url := js.Global.Get("URL").Call("createObjectURL", blob)
	defer js.Global.Get("URL").Call("revokeObjectURL", url)
	a.Set("href", url)
	a.Set("download", filename)
	a.Call("click")
}
