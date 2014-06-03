package models

import (
	"encoding/json"
	"errors"
	"reflect"
	"time"
)

type SyncRequest struct {
	Type        string
	LastUpdated time.Time
	Values      []interface{}
}

type syncRequestRaw struct {
	Type        string
	LastUpdated time.Time
	RawValues   json.RawMessage `json:"Values"`
}

func toValues(t string, raw json.RawMessage) (values []interface{}, err error) {
	switch t {
	case reflect.TypeOf((*ItemData)(nil)).Elem().Name():
		values, err = toItemData(raw)
	default:
		err = errors.New("SyncRequest.UnmarshalJSON: unknown type")
	}
	return
}

func toItemData(raw json.RawMessage) ([]interface{}, error) {
	values := []*ItemData{}
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	result := make([]interface{}, len(values))
	for i, v := range values {
		if !v.IsValid() {
			return nil, errors.New("protocol: invalid item")
		}
		result[i] = v
	}
	return result, nil
}

func (s *SyncRequest) UnmarshalJSON(b []byte) (err error) {
	raw := syncRequestRaw{}
	if err = json.Unmarshal(b, &raw); err != nil {
		return
	}
	s.Type = raw.Type
	s.LastUpdated = raw.LastUpdated
	s.Values, err = toValues(s.Type, raw.RawValues)
	return
}

type SyncResponse struct {
	Type        string
	LastUpdated time.Time
	Values      []interface{}
}

type syncResponseRaw struct {
	Type        string
	LastUpdated time.Time
	RawValues   json.RawMessage `json:"Values"`
}

func (s *SyncResponse) UnmarshalJSON(b []byte) (err error) {
	raw := syncResponseRaw{}
	if err = json.Unmarshal(b, &raw); err != nil {
		return
	}
	s.Type = raw.Type
	s.LastUpdated = raw.LastUpdated
	s.Values, err = toValues(s.Type, raw.RawValues)
	return
}
