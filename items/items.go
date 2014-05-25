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
	SetEdittingItem(id uuid.UUID)
	PrintItems(ids []uuid.UUID)
	PrintItem(data models.ItemData)
	PrintYearMonths([]date.Date)
	OnInit(items *Items)
}

type Mode int

const (
	ModeAll Mode = iota
	ModeYearMonth
)

// TODO: Should this have 'mode'?
type Items struct {
	items       map[uuid.UUID]*Item
	view        ItemsView
	storage     Storage
	mode        Mode
	yearMonth   date.Date
	editingItem *Item
}

func New(view ItemsView, storage Storage) *Items {
	return &Items{
		items:   map[uuid.UUID]*Item{},
		view:    view,
		storage: storage,
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
			item.print()
			continue
		}
		item := &Item{
			data:    d,
			view:    i.view,
			storage: i.storage,
		}
		i.items[id] = item
		item.print()
	}
	i.printYearMonths()
}

func (i *Items) OnInitialLoaded(vals []interface{}) {
	i.OnLoaded(vals)
	i.view.OnInit(i)
	i.createEdittingItem()
}

func (i *Items) createEdittingItem() error {
	item := newItem(i.view, i.storage)
	i.editingItem = item
	id := item.data.Meta.ID
	i.items[id] = item
	i.view.SetEdittingItem(id)
	if err := i.Print(id); err != nil {
		return err
	}
	return nil
}

func (i *Items) Print(id uuid.UUID) error {
	item := i.get(id)
	if item == nil {
		return errors.New("Items.Print: item not found")
	}
	item.print()
	return nil
}

func (i *Items) UpdateDate(id uuid.UUID, date date.Date) error {
	item := i.get(id)
	if item == nil {
		return errors.New("Items.UpdateDate: item not found")
	}
	item.updateDate(date)
	return nil
}

func (i *Items) UpdateSubject(id uuid.UUID, subject string) error {
	item := i.get(id)
	if item == nil {
		return errors.New("Items.UpdateSubject: item not found")
	}
	item.updateSubject(subject)
	return nil
}

func (i *Items) UpdateAmount(id uuid.UUID, amount models.MoneyAmount) error {
	item := i.get(id)
	if item == nil {
		return errors.New("Items.UpdateAmount: item not found")
	}
	item.updateAmount(amount)
	return nil
}

func (i *Items) Save(id uuid.UUID) error {
	item := i.get(id)
	if item == nil {
		return errors.New("Items.Save: item not found")
	}
	// TODO: Validation here
	if err := item.save(); err != nil {
		return err
	}
	if i.editingItem == item {
		i.createEdittingItem()
	}
	i.printItems()
	i.printYearMonths()
	return nil
}

func (i *Items) Destroy(id uuid.UUID) error {
	item := i.get(id)
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
	i.printItems()
}

func (i *Items) printItems() {
	switch i.mode {
	case ModeAll:
		i.printAllItems()
	case ModeYearMonth:
		i.printYearMonthItems()
	}
}

/*type sortItemsByDate []*items.Item

func (t sortItemsByDate) Len() int {
	return len(([]*items.Item)(t))
}

func (t sortItemsByDate) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t sortItemsByDate) Less(i, j int) bool {
	return false
}*/

func (i *Items) printAllItems() {
	ids := []uuid.UUID{}
	for _, item := range i.items {
		if item.data.Meta.IsDeleted {
			continue
		}
		if item == i.editingItem {
			continue
		}
		ids = append(ids, item.data.Meta.ID)
	}
	i.view.PrintItems(ids)
	for _, id := range ids {
		i.get(id).print()
	}
}

func (i *Items) printYearMonthItems() {
	ym := i.yearMonth
	ids := []uuid.UUID{}
	for _, item := range i.items {
		if item.data.Meta.IsDeleted {
			continue
		}
		if item == i.editingItem {
			continue
		}
		d := item.data.Date
		if d.Year() != ym.Year() || d.Month() != ym.Month() {
			continue
		}
		ids = append(ids, item.data.Meta.ID)
	}
	i.view.PrintItems(ids)
	for _, id := range ids {
		i.get(id).print()
	}
}

func (i *Items) get(id uuid.UUID) *Item {
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
	i.view.PrintYearMonths(result)
}
