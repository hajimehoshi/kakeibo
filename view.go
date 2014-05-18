package main

import (
	"errors"
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/kakeibo/models"
	"github.com/hajimehoshi/kakeibo/uuid"
	"strconv"
)

const (
	// TODO: Rename data-id -> data-models-id?
	datasetAttrID = "id"
	datasetAttrKey = "key"
)

func getIDElement(e js.Object) js.Object {
	for {
		attr := e.Get("dataset").Get(datasetAttrID)
		if !attr.IsUndefined() {
			return e
			/*str := attr.Str()
			id, err := uuid.ParseString(str)
			if err != nil {
				return uuid.UUID{}, err
			}
			return id, nil*/
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
		return uuid.UUID{}, errors.New("view: element not found")
	}
	str := e2.Get("dataset").Get(datasetAttrID).Str()
	id, err := uuid.ParseString(str)
	if err != nil {
		return uuid.UUID{}, err
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

func (p *HTMLView) PrintItem(data models.ItemData) {
	document := js.Global.Get("document")
	id := data.Meta.ID
	// TODO: Escape
	query := fmt.Sprintf("*[data-%s=\"%s\"]", datasetAttrID, id.String())
	elements := document.Call("querySelectorAll", query)
	for i := 0; i < elements.Length(); i++ {
		e := elements.Index(i)
		printValueAt(e, "date", data.Date.String())
		printValueAt(e, "subject", data.Subject)
		printValueAt(e, "amount", strconv.Itoa(int(data.Amount)))
	}
}

func (p *HTMLView) SetIDsToItemTable(ids []uuid.UUID) {
	document := js.Global.Get("document")
	table := document.Call("getElementById", "table_items")
	tbody := table.Call("getElementsByTagName", "tbody").Index(0)
	for tbody.Call("hasChildNodes").Bool() {
		tbody.Call("removeChild", tbody.Get("lastChild"))
	}
	// TODO: Sort here!
	for _, id := range ids {
		p.AddIDToItemTable(id)
	}
}

func (p *HTMLView) isEditting(id uuid.UUID) bool {
	document := js.Global.Get("document")
	form := document.Call("getElementById", "form_item")
	i := form.Get("dataset").Get("data-" + datasetAttrID).Str()
	return i == id.String()
}

func (p *HTMLView) AddIDToItemTable(id uuid.UUID) {
	if p.isEditting(id) {
		return
	}

	document := js.Global.Get("document")
	table := document.Call("getElementById", "table_items")
	query := fmt.Sprintf("tr[data-%s=\"%s\"]", datasetAttrID, id.String())
	if 1 <= table.Call("querySelectorAll", query).Length() {
		return
	}
	tr := document.Call("createElement", "tr")
	tr.Get("dataset").Set(datasetAttrID, id.String())

	td := document.Call("createElement", "td")
	td.Get("dataset").Set(datasetAttrKey, "date")
	tr.Call("appendChild", td)

	td = document.Call("createElement", "td")
	td.Get("dataset").Set(datasetAttrKey, "subject")
	tr.Call("appendChild", td)

	td = document.Call("createElement", "td")
	td.Get("dataset").Set(datasetAttrKey, "amount")
	tr.Call("appendChild", td)

	a := document.Call("createElement", "a")
	a.Set("textContent", "Delete")
	a.Call("setAttribute", "href", "")
	td = document.Call("createElement", "td")
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
