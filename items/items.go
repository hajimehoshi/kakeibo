package items

import (
	"errors"
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/models"
	"github.com/hajimehoshi/kakeibo/uuid"
	"reflect"
	"sort"
)

type Storage interface {
	Save(interface{}) error
}

type ItemsView interface {
	PrintItems(ids []uuid.UUID)
	PrintYearMonths([]date.Date)
	OnInit(items *Items)
}

type ItemView interface {
	PrintItem(data models.ItemData)
}

type Mode int

const (
	ModeAll Mode = iota
	ModeYearMonth
)

// TODO: Should this have 'mode'?
type Items struct {
	items     map[uuid.UUID]*Item
	itemsView ItemsView
	itemView  ItemView
	storage   Storage
	mode      Mode
	yearMonth date.Date
}

func New(itemsView ItemsView, itemView ItemView, storage Storage) *Items {
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
	i.printYearMonths()
}

func (i *Items) OnInitialLoaded(vals []interface{}) {
	i.OnLoaded(vals)
	i.itemsView.OnInit(i)
}

// TODO: Make this private
func (i *Items) New() *Item {
	item := NewItem(i.itemView, i.storage)
	i.items[item.data.Meta.ID] = item
	i.printYearMonths()
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
	i.printYearMonths()
	return nil
}


func (i *Items) Destroy(id uuid.UUID) error {
	item := i.Get(id)
	if item == nil {
		return errors.New("Items.Save: item not found")
	}
	meta := item.data.Meta
	meta.LastUpdated = models.UnixTime(0)
	meta.IsDeleted = true
	item.data = &models.ItemData{Meta: meta}
	if err := item.save(); err != nil {
		return err
	}
	i.printYearMonths()
	return nil
}

func (i *Items) UpdateMode(mode Mode, ym date.Date) {
	i.mode = mode
	i.yearMonth = ym
	switch i.mode {
	case ModeAll:
		i.printAllItems()
	case ModeYearMonth:
		i.printYearMonthItems()
	}
}

func (i *Items) printAllItems() {
	ids := []uuid.UUID{}
	for _, item := range i.items {
		if item.data.Meta.IsDeleted {
			continue
		}
		ids = append(ids, item.data.Meta.ID)
	}
	i.itemsView.PrintItems(ids)
	for _, id := range ids {
		i.Get(id).Print()
	}
}

func (i *Items) printYearMonthItems() {
	ym := i.yearMonth
	ids := []uuid.UUID{}
	for _, item := range i.items {
		if item.data.Meta.IsDeleted {
			continue
		}
		d := item.data.Date
		if d.Year() != ym.Year() || d.Month() != ym.Month() {
			continue
		}
		ids = append(ids, item.data.Meta.ID)
	}
	i.itemsView.PrintItems(ids)
	for _, id := range ids {
		i.Get(id).Print()
	}
}

func (i *Items) Get(id uuid.UUID) *Item {
	if item, ok := i.items[id]; ok {
		return item
	}
	return nil
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

func (i *Items) printYearMonths() {
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
	s := sortDateDesc(result)
	sort.Sort(s)
	i.itemsView.PrintYearMonths(result)
}
