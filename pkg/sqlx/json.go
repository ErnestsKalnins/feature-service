// Package sqlx provides convenience methods for dealing with SQL databases.
package sqlx

import (
	"encoding/json"
	"fmt"
)

// JSONArray allows scanning SQLite JSON functions that result with an array into
// a native Go slice.
type JSONArray[T bool | float64 | string] []T

// Scan implements the sql.Scanner interface.
func (a *JSONArray[T]) Scan(src any) error {
	// SQLite JSON functions generally return JSON as string.
	switch v := src.(type) {
	case nil:
		return nil
	case string:
		return json.Unmarshal([]byte(v), a)
	default:
		return fmt.Errorf("cannot scan %T into %T", src, a)
	}
}
