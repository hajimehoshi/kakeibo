package view

import (
	"errors"
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/items"
	"github.com/hajimehoshi/kakeibo/models"
	"github.com/hajimehoshi/kakeibo/uuid"
	"reflect"
	"strconv"
)

type Items interface {
	New() uuid.UUID
	UpdateDate(id uuid.UUID, date date.Date) error
	UpdateSubject(id uuid.UUID, subject string) error
	UpdateAmount(id uuid.UUID, amount models.MoneyAmount) error
	Save(id uuid.UUID) error
	Destroy(id uuid.UUID) error
	UpdateMode(mode items.Mode, ym date.Date)
}

func printError(val interface{}) {
	js.Global.Get("console").Call("error", val)
}

// TODO: I18N

const (
	// TODO: Rename data-id -> data-models-id?
	datasetAttrID  = "id"
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

func addEventListeners(items Items, form js.Object) {
	inputDate := form.Call("querySelector", "input[name=Date]")
	inputDate.Set("onchange", func(e js.Object) {
		id, err := getIDFromElement(e.Get("target"))
		if err != nil {
			printError(err.Error())
			return
		}
		dateStr := e.Get("target").Get("value").Str()
		d, err := date.ParseISO8601(dateStr)
		if err != nil {
			printError(err.Error())
			return
		}
		if err := items.UpdateDate(id, d); err != nil {
			printError(err.Error())
			return
		}
	})
	inputSubject := form.Call("querySelector", "input[name=Subject]")
	inputSubject.Set("onchange", func(e js.Object) {
		id, err := getIDFromElement(e.Get("target"))
		if err != nil {
			printError(err.Error())
			return
		}
		subject := e.Get("target").Get("value").Str()
		if err := items.UpdateSubject(id, subject); err != nil {
			printError(err.Error())
			return
		}
	})
	inputMoneyAmount := form.Call("querySelector", "input[name=Amount]")
	inputMoneyAmount.Set("onchange", func(e js.Object) {
		id, err := getIDFromElement(e.Get("target"))
		if err != nil {
			printError(err.Error())
			return
		}
		amount := models.MoneyAmount(e.Get("target").Get("value").Int())
		if err := items.UpdateAmount(id, amount); err != nil {
			printError(err.Error())
			return
		}
	})
}

type HTMLView struct {
	items Items
	queue []func()
}

func empty(e js.Object) {
	for e.Call("hasChildNodes").Bool() {
		e.Call("removeChild", e.Get("lastChild"))
	}
}

func NewHTMLView() *HTMLView {
	v := &HTMLView{
		queue: []func(){},
	}
	document := js.Global.Get("document")
	form := document.Call("getElementById", "form_item")
	form.Set("onsubmit", v.onSubmit)

	js.Global.Get("window").Set("onhashchange", v.onHashChange)
	js.Global.Get("window").Call("onhashchange")

	return v
}

func (v *HTMLView) onHashChange(e js.Object) {
	hash := js.Global.Get("location").Get("hash").Str()
	// Remove the initial '#'
	if 1 <= len(hash) {
		hash = hash[1:]
	}
	if hash == "" {
		href := js.Global.Get("location").Get("href").Str()
		if 0 < len(href) && href[len(href)-1] == '#' {
			href = href[:len(href)-2]
			js.Global.Get("history").Call(
				"replaceState", "", "", href)
		}
		v.UpdateMode(items.ModeAll, date.Date(0))
		return
	}
	ym, err := date.ParseISO8601(hash + "-01")
	if err != nil {
		printError(err.Error())
		return
	}
	v.UpdateMode(items.ModeYearMonth, ym)
}

func (v *HTMLView) onSubmit(e js.Object) {
	e.Call("preventDefault")
	if !v.isInited() {
		return
	}
	form := e.Get("target")
	id, err := getIDFromElement(form)
	if err != nil {
		printError(err.Error())
		return
	}
	if err := v.items.Save(id); err != nil {
		printError(err.Error())
		return
	}

	// FIXME: Before saving an item, the form's item should be
	// changed?
	id = v.items.New()
	form.Get("dataset").Set(datasetAttrID, id.String())
}

func (v *HTMLView) isInited() bool {
	return v.items != nil
}

func (v *HTMLView) UpdateMode(mode items.Mode, ym date.Date) {
	if !v.isInited() {
		v.queue = append(v.queue, func() {
			v.UpdateMode(mode, ym)
		})
		return
	}
	v.items.UpdateMode(mode, ym)
}

func (v *HTMLView) PrintItems(ids []uuid.UUID) {
	document := js.Global.Get("document")
	table := document.Call("getElementById", "table_items")
	tbody := table.Call("getElementsByTagName", "tbody").Index(0)
	empty(tbody)
	for _, id := range ids {
		v.addIDToItemTable(id)
	}
}

func (v *HTMLView) OnInit(items *items.Items) {
	v.items = items
	for _, f := range v.queue {
		f()
	}
	v.queue = []func(){}

	document := js.Global.Get("document")
	form := document.Call("getElementById", "form_item")
	addEventListeners(items, form)
	id := items.New()
	form.Get("dataset").Set(datasetAttrID, id.String())
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

func (v *HTMLView) isEditting(id uuid.UUID) bool {
	document := js.Global.Get("document")
	form := document.Call("getElementById", "form_item")
	i := form.Get("dataset").Get(datasetAttrID).Str()
	return i == id.String()
}

func (v *HTMLView) addIDToItemTable(id uuid.UUID) {
	if v.isEditting(id) {
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
	a.Set("onclick", v.onClickToDelete)
	tr.Call("appendChild", td)

	tbody := table.Call("getElementsByTagName", "tbody").Index(0)
	tbody.Call("appendChild", tr)
}

func (v *HTMLView) onClickToDelete(e js.Object) {
	e.Call("preventDefault")
	id, err := getIDFromElement(e.Get("target"))
	if err != nil {
		printError(err.Error())
		return
	}
	// TODO: Confirming if needed.
	if err := v.items.Destroy(id); err != nil {
		printError(err.Error())
		return
	}
}
