package packageInfo

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type PackageAttribute struct {
	Key      string `json:"key"`
	Fixed    bool   `json:"fixed"`
	Value    string `json:"value"`
	Hidden   bool   `json:"hidden"`
	Category string `json:"category"`
	DataType string `json:"dataType"`
}

type PackageAttributes []PackageAttribute

// Linter complains about mixing value and pointer PackageAttribute receivers
// But I could only get this to work for inserts and selects if Value() has
// a non-pointer receiver and Scan() has a pointer receiver.

func (a PackageAttributes) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *PackageAttributes) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion PackageAttributes to []byte failed")
	}

	return json.Unmarshal(b, &a)
}
