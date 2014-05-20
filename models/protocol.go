package models

import (
	"encoding/json"
	"errors"
	"reflect"
)

type SyncRequest struct {
	Type        string
	LastUpdated UnixTime
	Values      []interface{}
}

type syncRequestRaw struct {
	Type        string
	LastUpdated UnixTime
	RawValues   json.RawMessage `json:"Values"`
}

func (s *SyncRequest) UnmarshalJSON(b []byte) error {
	raw := syncRequestRaw{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	s.Type = raw.Type
	s.LastUpdated = raw.LastUpdated
	switch s.Type {
	case reflect.TypeOf(ItemData{}).Name():
		values := []*ItemData{}
		if err := json.Unmarshal(raw.RawValues, &values); err != nil {
			return err
		}
		for _, v := range values {
			s.Values = append(s.Values, v)
		}
	default:
		return errors.New("SyncRequest.UnmarshalJSON: unknown type")
	}
	return nil
}

// TODO: Refactoring
type SyncResponse struct {
	Type        string
	LastUpdated UnixTime
	Values      []interface{}
}

type syncResponseRaw struct {
	Type        string
	LastUpdated UnixTime
	RawValues   json.RawMessage `json:"Values"`
}

func (s *SyncResponse) UnmarshalJSON(b []byte) error {
	raw := syncResponseRaw{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	s.Type = raw.Type
	s.LastUpdated = raw.LastUpdated
	switch s.Type {
	case reflect.TypeOf(ItemData{}).Name():
		values := []*ItemData{}
		if err := json.Unmarshal(raw.RawValues, &values); err != nil {
			return err
		}
		for _, v := range values {
			s.Values = append(s.Values, v)
		}
	default:
		return errors.New("SyncResponse.UnmarshalJSON: unknown type")
	}
	return nil
}
