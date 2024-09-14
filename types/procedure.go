package types

import (
	"encoding/json"
	"reflect"
)

type JsonSerializable interface {
	json.Marshaler
	json.Unmarshaler
}

// No-op type to represent a procedure that doesn't return any response or take any payload
type Void struct{}

func (v Void) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

func (v *Void) UnmarshalJSON(data []byte) error {
	if string(data) != "" && string(data) != "null" {
		return &json.UnsupportedValueError{
			Value: reflect.ValueOf(data),
			Str:   string(data),
		}
	}
	return nil
}

var _ JsonSerializable = (*Void)(nil)
