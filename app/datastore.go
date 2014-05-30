package index

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"github.com/hajimehoshi/kakeibo/models"
	"github.com/hajimehoshi/kakeibo/uuid"
	"errors"
	"reflect"
	"time"
)

const (
	kindItems = "Items"
)

var (
	rootKeyStringID = reflect.TypeOf((*models.ItemData)(nil)).Elem().Name()
)

type ItemDatastore struct {
	context appengine.Context
	userID  string
	rootKey *datastore.Key
}

func NewItemDatastore(context appengine.Context, userID string) *ItemDatastore {
	rootKey := datastore.NewKey(
		context,
		kindItems,
		rootKeyStringID,
		0,
		nil)
	return &ItemDatastore{
		context: context,
		userID:  userID,
		rootKey: rootKey,
	}
}

func (d *ItemDatastore) datastoreKey(id uuid.UUID) *datastore.Key {
	return datastore.NewKey(
		d.context,
		kindItems,
		id.String(),
		0,
		d.rootKey)
}

func (d *ItemDatastore) Put(
	lastUpdated models.UnixTime,
	reqItems []*models.ItemData) (now models.UnixTime, err error) {
	now = models.UnixTime(time.Now().Unix())
	if now < lastUpdated {
		err = errors.New("last-updated is too new")
		return 
	}
	f := func(c appengine.Context) error {
		itemsToPut := []*models.ItemData{}
		for _, item := range reqItems {
			id := item.Meta.ID
			var existingData models.ItemData
			key := d.datastoreKey(id)
			err := datastore.Get(c, key, &existingData)
			switch err {
			case nil:
				if d.userID != existingData.Meta.UserID {
					e := fmt.Sprintf(
						"invalid UUID: %s",
						id.String())
					return errors.New(e)
				}
				if existingData.Meta.LastUpdated >= lastUpdated {
					continue
				}
			case datastore.ErrNoSuchEntity:
			default:
				return err
			}
			item.Meta.LastUpdated = now
			item.Meta.UserID = d.userID
			itemsToPut = append(itemsToPut, item)
		}
		keys := make([]*datastore.Key, len(itemsToPut))
		for i, item := range itemsToPut {
			key := d.datastoreKey(item.Meta.ID)
			keys[i] = key
		}
		_, err := datastore.PutMulti(c, keys, itemsToPut)
		return err
	}
	err = datastore.RunInTransaction(d.context, f, nil)
	return 
}

func (d *ItemDatastore) Get(
	lastUpdated models.UnixTime) (items []*models.ItemData, err error) {
	q := datastore.NewQuery(kindItems).
		Ancestor(d.rootKey).
		Filter("Meta.LastUpdated >=", lastUpdated).
		Filter("Meta.UserID =", d.userID)
	items = []*models.ItemData{}
	_, err = q.GetAll(d.context, &items)
	return
}



