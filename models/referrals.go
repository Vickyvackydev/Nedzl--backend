package models

import (
	"encoding/json"
	"errors"
)

func (r *ReferedBy) Scan(value interface{}) error {
	// Implementation for scanning the value into the ReferedBy struct

	if value == nil {

		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Failed to scan ReferredBy")
	}
	return json.Unmarshal(bytes, r)
}

func (r ReferedBy) Value() (interface{}, error) {
	// Implementation for converting the ReferedBy struct to a database value
	return json.Marshal(r)
}
