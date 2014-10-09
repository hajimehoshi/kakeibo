package items

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/hajimehoshi/kakeibo/date"
	"github.com/hajimehoshi/kakeibo/models"
	"github.com/hajimehoshi/kakeibo/uuid"
	"reflect"
	"sort"
	"time"
)

type Storage interface {
	// TODO: This takes interface{} because of the IndexedDB. This can be
	// fixed.
	Save(interface{}) error
}

type ItemsView interface {
	SetEditingItem(id uuid.UUID)
	PrintTitle(title string)
	PrintItems(ids []uuid.UUID)
	PrintItemsAndTotal(ids []uuid.UUID, total int)
	PrintItem(data models.ItemData)
	PrintYearMonths([]date.Date)
	Download(b []byte, filename string)
}

type Mode int

const (
	ModeTop Mode = iota
	ModeYearMonth
)

// TODO: Should this have 'mode'?
type Items struct {
	items       map[uuid.UUID]*models.ItemData
	view        ItemsView
	storage     Storage
	mode        Mode
	yearMonth   date.Date
	editingItem *models.ItemData
}

func New(view ItemsView, storage Storage) *Items {
	items := &Items{
		items:   map[uuid.UUID]*models.ItemData{},
		view:    view,
		storage: storage,
	}
	items.createEditingItem(date.Today())
	return items
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
			*item = *d
			i.printItem(item)
			continue
		}
		i.items[id] = d
	}
	i.printYearMonths()
	i.printItems()
}

func (i *Items) createEditingItem(date date.Date) error {
	item := &models.ItemData{
		Meta: models.Meta{ID: uuid.Generate()},
	}
	item.Date = date
	i.editingItem = item
	id := item.Meta.ID
	i.items[id] = item
	i.view.SetEditingItem(id)
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
	if i.view == nil {
		return nil
	}
	i.view.PrintItem(*item)
	return nil
}

func (i *Items) printItem(item *models.ItemData) {
	if i.view == nil {
		return
	}
	i.view.PrintItem(*item)
}

func (i *Items) saveItem(item *models.ItemData) error {
	// TODO: Confine calling IsValid here
	if !item.IsValid() {
		return errors.New("Items.saveItem: invalid data")
	}
	item.Meta.LastUpdated = time.Time{}
	if i.storage == nil {
		return nil
	}
	err := i.storage.Save(item) //gopherjs:blocking
	if err != nil {
		return err
	}
	return nil
}

func (i *Items) UpdateDate(id uuid.UUID, date date.Date) error {
	item := i.get(id)
	if item == nil {
		return errors.New("Items.UpdateDate: item not found")
	}
	item.Date = date
	i.printItem(item)
	return nil
}

func (i *Items) UpdateSubject(id uuid.UUID, subject string) error {
	item := i.get(id)
	if item == nil {
		return errors.New("Items.UpdateSubject: item not found")
	}
	item.Subject = subject
	i.printItem(item)
	return nil
}

func (i *Items) UpdateAmount(id uuid.UUID, amount int32) error {
	item := i.get(id)
	if item == nil {
		return errors.New("Items.UpdateAmount: item not found")
	}
	item.Amount = amount
	i.printItem(item)
	return nil
}

func (i *Items) Save(id uuid.UUID) error {
	item := i.get(id)
	if item == nil {
		return errors.New("Items.Save: item not found")
	}
	if err := i.saveItem(item); err != nil {
		return err
	}
	if i.editingItem == item {
		i.createEditingItem(item.Date)
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
	item.Destroy()
	if err := i.saveItem(item); err != nil {
		return err
	}
	i.printItem(item)
	i.printItems()
	i.printYearMonths()
	return nil
}

func (i *Items) title() string {
	switch i.mode {
	case ModeTop:
		return ""
	case ModeYearMonth:
		ym := i.yearMonth
		return fmt.Sprintf("%04d-%02d", ym.Year(), ym.Month())
	}
	panic("not reach")
}

func (i *Items) UpdateMode(mode Mode, ym date.Date) {
	i.mode = mode
	i.yearMonth = ym
	i.view.PrintTitle(i.title())
	i.printItems()
}

func (i *Items) printItems() {
	switch i.mode {
	case ModeTop:
		i.printNoItems()
	case ModeYearMonth:
		i.printYearMonthItems()
	}
}

type sortItemsByDate struct {
	items *Items
	ids   []uuid.UUID
}

func (s sortItemsByDate) Len() int {
	return len(([]uuid.UUID)(s.ids))
}

func (s sortItemsByDate) Swap(i, j int) {
	s.ids[i], s.ids[j] = s.ids[j], s.ids[i]
}

func (s sortItemsByDate) Less(i, j int) bool {
	i1 := s.items.get(s.ids[i])
	i2 := s.items.get(s.ids[j])
	if i1.Date != i2.Date {
		return i1.Date < i2.Date
	}
	if i1.Subject != i2.Subject {
		return i1.Subject < i2.Subject
	}
	return i1.Amount < i2.Amount
}

func (i *Items) printNoItems() {
	i.view.PrintItems([]uuid.UUID{})
}

func (i *Items) printYearMonthItems() {
	ym := i.yearMonth
	ids := []uuid.UUID{}
	total := 0
	for _, item := range i.items {
		if item.Meta.IsDeleted {
			continue
		}
		if item == i.editingItem {
			continue
		}
		d := item.Date
		if d.Year() != ym.Year() || d.Month() != ym.Month() {
			continue
		}
		ids = append(ids, item.Meta.ID)
		total += int(item.Amount)
	}
	s := sortItemsByDate{i, ids}
	sort.Sort(s)
	i.view.PrintItemsAndTotal(ids, total)
	for _, id := range ids {
		i.printItem(i.get(id))
	}
}

func (i *Items) get(id uuid.UUID) *models.ItemData {
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
		if item.Meta.IsDeleted {
			continue
		}
		if item == i.editingItem {
			continue
		}
		d := item.Date
		y := d.Year()
		m := d.Month()
		yms[date.New(y, m, 1)] = struct{}{}
	}

	result := make([]date.Date, 0, len(yms))
	for ym, _ := range yms {
		result = append(result, ym)
	}
	s := sortDateDesc(result)
	sort.Sort(s)
	i.view.PrintYearMonths(result)
}

func (i *Items) DownloadCSV() error {
	// TODO: Refactoring
	ids := []uuid.UUID{}
	for _, item := range i.items {
		if item.Meta.IsDeleted {
			continue
		}
		if item == i.editingItem {
			continue
		}
		ids = append(ids, item.Meta.ID)
	}
	s := sortItemsByDate{i, ids}
	sort.Sort(s)

	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	for _, id := range ids {
		item := i.get(id)
		r := item.CSVRecord()
		w.Write(r)
	}
	w.Flush()
	i.view.Download(buf.Bytes(), "kakeibo.csv")
	return nil
}
