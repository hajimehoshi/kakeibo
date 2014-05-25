package main

import (
	"errors"
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/models"
	"github.com/hajimehoshi/kakeibo/uuid"
	"reflect"
	"sort"
	"strconv"
)

// TODO: I18N

const (
	// TODO: Rename data-id -> data-models-id?
	datasetAttrID  = "id"
	datasetAttrKey = "key"
)

type ViewMode int

const (
	ViewModeAll ViewMode = iota
	ViewModeYearMonth
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

func addEventListeners(items *Items, form js.Object) {
	inputDate := form.Call("querySelector", "input[name=Date]")
	inputDate.Set("onchange", func(e js.Object) {
		id, err := getIDFromElement(e.Get("target"))
		if err != nil {
			printError(err.Error())
			return
		}
		item := items.Get(id)

		dateStr := e.Get("target").Get("value").Str()
		d, err := date.ParseISO8601(dateStr)
		if err != nil {
			printError(err.Error())
			return
		}
		item.UpdateDate(d)
	})
	inputSubject := form.Call("querySelector", "input[name=Subject]")
	inputSubject.Set("onchange", func(e js.Object) {
		id, err := getIDFromElement(e.Get("target"))
		if err != nil {
			printError(err.Error())
			return
		}
		item := items.Get(id)

		subject := e.Get("target").Get("value").Str()
		item.UpdateSubject(subject)
	})
	inputMoneyAmount := form.Call("querySelector", "input[name=Amount]")
	inputMoneyAmount.Set("onchange", func(e js.Object) {
		id, err := getIDFromElement(e.Get("target"))
		if err != nil {
			printError(err.Error())
			return
		}
		item := items.Get(id)

		amount := e.Get("target").Get("value").Int()
		item.UpdateAmount(models.MoneyAmount(amount))
	})
}

type HTMLView struct {
	items     *Items
	queue     []func()
	mode      ViewMode
	yearMonth date.Date
}

func empty(e js.Object) {
	for e.Call("hasChildNodes").Bool() {
		e.Call("removeChild", e.Get("lastChild"))
	}
}

func NewHTMLView() *HTMLView {
	v := &HTMLView{
		items:     nil,
		queue:     []func(){},
		mode:      ViewModeAll,
		yearMonth: date.Date(0),
	}
	// TODO: ?
	document := js.Global.Get("document")
	form := document.Call("getElementById", "form_item")
	
	form.Set("onsubmit", func(e js.Object) {
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
		newItem := v.items.New()
		form.Get("dataset").Set(datasetAttrID, newItem.ID().String())
		newItem.Print()
	})

	return v
}

func (v *HTMLView) isInited() bool {
	return v.items != nil
}

func (v *HTMLView) UpdateMode(mode ViewMode, ym date.Date) {
	v.mode = mode
	v.yearMonth = ym
	if !v.isInited() {
		v.queue = append(v.queue, func() {
			v.UpdateMode(mode, ym)
		})
		return
	}
	v.PrintItems()
}

type sortItemsByDate []*Item

func (t sortItemsByDate) Len() int {
	return len(([]*Item)(t))
}

func (t sortItemsByDate) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t sortItemsByDate) Less(i, j int) bool {
	return false
}

func (v *HTMLView) PrintItems() {
	document := js.Global.Get("document")
	table := document.Call("getElementById", "table_items")
	tbody := table.Call("getElementsByTagName", "tbody").Index(0)
	empty(tbody)
	
	// FIXME: sort
	switch v.mode {
	case ViewModeAll:
		for _, i := range v.items.GetAll() {
			v.addIDToItemTable(i.ID())
			i.Print()
		}
	case ViewModeYearMonth:
		ym := v.yearMonth
		is := v.items.GetYearMonth(ym.Year(), ym.Month())
		for _, i := range is {
			v.addIDToItemTable(i.ID())
			i.Print()
		}
	}
}

func (v *HTMLView) OnInit(items *Items) {
	v.items = items
	items.PrintYearMonths()
	for _, f := range v.queue {
		f()
	}
	v.queue = []func(){}

	document := js.Global.Get("document")
	form := document.Call("getElementById", "form_item")
	addEventListeners(items, form)

	item := items.New()
	form.Get("dataset").Set(datasetAttrID, item.ID().String())
	item.Print()
}

type sortDateDesc []date.Date

func (s sortDateDesc) Len() int {
	return len(([]date.Date)(s))
}

func (s sortDateDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortDateDesc) Less(i, j int) bool {
	return s[i] > s[j]
}

func (v *HTMLView) PrintYearMonths(yms []date.Date) {
	document := js.Global.Get("document")
	ul := document.Call("getElementById", "year_months")
	empty(ul)
	s := sortDateDesc(yms)
	sort.Sort(s)
	for _, ym := range s {
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
	a.Set("onclick", v.clickLinkToDelete)
	tr.Call("appendChild", td)

	tbody := table.Call("getElementsByTagName", "tbody").Index(0)
	tbody.Call("appendChild", tr)
}

func (v *HTMLView) clickLinkToDelete(e js.Object) {
	e.Call("preventDefault")
	id, err := getIDFromElement(e.Get("target"))
	if err != nil {
		printError(err.Error())
		return
	}
	// TODO: Confirming if needed.
	item := v.items.Get(id)
	item.Destroy()
	e2 := getIDElement(e.Get("target"))
	if e2 != nil {
		e2.Get("parentNode").Call("removeChild", e2)
	}
}
