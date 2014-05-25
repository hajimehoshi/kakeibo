package main

import (
	"errors"
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/models"
	"github.com/hajimehoshi/kakeibo/uuid"
	"reflect"
	"time"
)

type Storage interface {
	Save(interface{}) error
}

type ItemsView interface {
	PrintItems()
	PrintYearMonths([]date.Date)
}

type ItemView interface {
	PrintItem(data models.ItemData)
	OnInit(items *Items)
}

type Item struct {
	data    *models.ItemData
	view    ItemView
	storage Storage
}

func NewItem(view ItemView, storage Storage) *Item {
	return &Item{
		data: &models.ItemData{
			Meta: models.NewMeta(),
			Date: date.Today(),
		},
		view:    view,
		storage: storage,
	}
}

func (i *Item) ID() uuid.UUID {
	return i.data.Meta.ID
}

func (i *Item) UpdateDate(date date.Date) {
	i.data.Date = date
	i.Print()
}

func (i *Item) UpdateSubject(subject string) {
	i.data.Subject = subject
	i.Print()
}

func (i *Item) UpdateAmount(amount models.MoneyAmount) {
	i.data.Amount = amount
	i.Print()
}

func (i *Item) Destroy() {
	meta := i.data.Meta
	meta.LastUpdated = models.UnixTime(0)
	meta.IsDeleted = true
	i.data = &models.ItemData{Meta: meta}
	i.save()
}

func (i *Item) Print() {
	if i.view == nil {
		return
	}
	i.view.PrintItem(*i.data)
}

func (i *Item) save() error {
	i.data.Meta.LastUpdated = models.UnixTime(0)
	i.Print()
	if i.storage == nil {
		return nil
	}
	return i.storage.Save(i.data)
}

type Items struct {
	items     map[uuid.UUID]*Item
	itemsView ItemsView
	itemView  ItemView
	storage   Storage
}

func NewItems(itemsView ItemsView, itemView ItemView, storage Storage) *Items {
	return &Items{
		items:     map[uuid.UUID]*Item{},
		itemsView: itemsView,
		itemView:  itemView,
		storage:   storage,
	}
}

func (i *Items) Type() reflect.Type {
	return reflect.TypeOf((*models.ItemData)(nil)).Elem()
}

func (i *Items) OnLoaded(vals []interface{}) {
	for _, v := range vals {
		d, ok := v.(*models.ItemData)
		if !ok {
			print("invalid data")
			return
		}
		id := d.Meta.ID
		if item, ok := i.items[id]; ok {
			*item.data = *d
			item.Print()
			continue
		}
		item := &Item{
			data:    d,
			view:    i.itemView,
			storage: i.storage,
		}
		i.items[id] = item
		item.Print()
	}
	i.PrintYearMonths()
}

func (i *Items) OnInitialLoaded(vals []interface{}) {
	i.OnLoaded(vals)
	i.itemView.OnInit(i)
}

func (i *Items) New() *Item {
	item := NewItem(i.itemView, i.storage)
	i.items[item.data.Meta.ID] = item
	i.PrintYearMonths()
	return item
}

func (i *Items) Save(id uuid.UUID) error {
	item := i.Get(id)
	if item == nil {
		return errors.New("Items.Save: item not found")
	}
	// TODO: Validation here
	if err := item.save(); err != nil {
		return err
	}
	i.itemsView.PrintItems()
	return nil
}

func (i *Items) Get(id uuid.UUID) *Item {
	if item, ok := i.items[id]; ok {
		return item
	}
	return nil
}

func (i *Items) GetAll() []*Item {
	result := []*Item{}
	for _, item := range i.items {
		if item.data.Meta.IsDeleted {
			continue
		}
		result = append(result, item)
	}
	return result
}

func (i *Items) GetYearMonth(year int, month time.Month) []*Item {
	result := []*Item{}
	for _, item := range i.items {
		if item.data.Meta.IsDeleted {
			continue
		}
		d := item.data.Date
		if d.Year() != year || d.Month() != month {
			continue
		}
		result = append(result, item)
	}
	return result
}

func (i *Items) PrintYearMonths() {
	yms := map[date.Date]struct{}{}
	for _, item := range i.items {
		if item.data.Meta.IsDeleted {
			continue
		}
		d := item.data.Date
		y := d.Year()
		m := d.Month()
		yms[date.New(y, m, 1)] = struct{}{}
	}

	result := []date.Date{}
	for ym, _ := range yms {
		result = append(result, ym)
	}
	i.itemsView.PrintYearMonths(result)
}
