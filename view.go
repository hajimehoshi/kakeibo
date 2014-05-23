package main

import (
	"errors"
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/models"
	"github.com/hajimehoshi/kakeibo/uuid"
	"reflect"
	//"sort"
	"strconv"
)

// TODO: I18N

const (
	// TODO: Rename data-id -> data-models-id?
	datasetAttrID = "id"
	datasetAttrKey = "key"
)

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
		attr := e.Get("dataset").Get(datasetAttrID)
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
		return uuid.Zero, errors.New("view: element not found")
	}
	str := e2.Get("dataset").Get(datasetAttrID).Str()
	id, err := uuid.ParseString(str)
	if err != nil {
		return uuid.Zero, err
	}
	return id, nil
}

func printValueAt(e js.Object, name string, value string) {
	targets := []js.Object{}
	if e.Get("name").Str() == name {
		targets = append(targets, e)
	}
	// TODO: Escape
	query := fmt.Sprintf("*[name=\"%s\"]", name)
	es := e.Call("querySelectorAll", query)
	for i := 0; i < es.Length(); i++ {
		targets = append(targets, es.Index(i))
	}

	if e.Get("dataset").Get(datasetAttrKey).Str() == name {
		targets = append(targets, e)
	}
	// TODO: Escape
	query = fmt.Sprintf("*[data-%s=\"%s\"]", datasetAttrKey, name)
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

type HTMLView struct{}

func empty(e js.Object) {
	for e.Call("hasChildNodes").Bool() {
		e.Call("removeChild", e.Get("lastChild"))
	}
}

func (p *HTMLView) OnInit(items *Items) {
	document := js.Global.Get("document")
	table := document.Call("getElementById", "table_items")
	tbody := table.Call("getElementsByTagName", "tbody").Index(0)
	empty(tbody)
	for _, i := range items.GetAll() {
		p.AddIDToItemTable(i.ID())
		i.Print()
	}
	items.PrintYearMonths()
}

type sortDesc []date.Date

func (s sortDesc) Len() int {
	return len(([]date.Date)(s))
}

func (s sortDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortDesc) Less(i, j int) bool {
	return s[i] > s[j]
}

func (p *HTMLView) PrintYearMonths(yms []date.Date) {
	document := js.Global.Get("document")
	ul := document.Call("getElementById", "year_months")
	empty(ul)
	/*sort.Sort(yms)
	for _, ym := range yms {
		
	}*/
}

func (p *HTMLView) PrintItem(data models.ItemData) {
	document := js.Global.Get("document")
	id := data.Meta.ID
	// TODO: Escape
	query := fmt.Sprintf("*[data-%s=\"%s\"]", datasetAttrID, id.String())
	elements := document.Call("querySelectorAll", query)
	for i := 0; i < elements.Length(); i++ {
		e := elements.Index(i)
		printValueAt(e, "Date", data.Date.String())
		printValueAt(e, "Subject", data.Subject)
		printValueAt(e, "Amount", strconv.Itoa(int(data.Amount)))
	}
}

func (p *HTMLView) isEditting(id uuid.UUID) bool {
	document := js.Global.Get("document")
	form := document.Call("getElementById", "form_item")
	i := form.Get("dataset").Get(datasetAttrID).Str()
	return i == id.String()
}

func (p *HTMLView) AddIDToItemTable(id uuid.UUID) {
	if p.isEditting(id) {
		return
	}
	t := reflect.TypeOf((*models.ItemData)(nil)).Elem()

	document := js.Global.Get("document")
	table := document.Call("getElementById", "table_items")
	query := fmt.Sprintf("tr[data-%s=\"%s\"]", datasetAttrID, id.String())
	if 1 <= table.Call("querySelectorAll", query).Length() {
		return
	}
	tr := document.Call("createElement", "tr")
	tr.Get("dataset").Set(datasetAttrID, id.String())

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type == reflect.TypeOf((*models.Meta)(nil)).Elem() {
			continue
		}
		td := document.Call("createElement", "td")
		td.Get("dataset").Set(datasetAttrKey, f.Name)
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
	a.Set("onclick", clickLinkToDelete)
	tr.Call("appendChild", td)

	tbody := table.Call("getElementsByTagName", "tbody").Index(0)
	tbody.Call("appendChild", tr)
}

func clickLinkToDelete(e js.Object) {
	e.Call("preventDefault")
	id, err := getIDFromElement(e.Get("target"))
	if err != nil {
		printError(err.Error())
		return
	}
	// TODO: Confirming if needed.
	item := items.Get(id)
	item.Destroy()
	e2 := getIDElement(e.Get("target")) 
	if e2 != nil {
		e2.Get("parentNode").Call("removeChild", e2)
	}
}
