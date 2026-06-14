package enum

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// This file holds the StrictEnum delegate helpers for ~string enum types.
//
// The enum type stays a bare string, so JSON and SQL are native. Each helper
// validates membership at the boundary: unknown values are rejected automatically,
// in both directions (Scan/Unmarshal in, Value/Marshal out). No reflect is used.

// ScanString implements sql.Scanner for a ~string enum type and validates membership.
// It accepts a string or []byte source; nil yields ErrNullValue and any other type
// yields ErrUnsupportedType.
func ScanString[T ~string](set *Set[T], dst *T, src any) error {
	switch value := src.(type) {
	case string:
		return assignString(set, dst, value)
	case []byte:
		return assignString(set, dst, string(value))
	case nil:
		return ErrNullValue
	default:
		return fmt.Errorf("%w: %T", ErrUnsupportedType, src)
	}
}

// assignString validates text against the set and, on success, stores it into dst.
func assignString[T ~string](set *Set[T], dst *T, text string) error {
	value := T(text)

	if !set.Has(value) {
		return fmt.Errorf("%w: %q", ErrInvalidValue, text)
	}

	*dst = value

	return nil
}

// ValueString implements driver.Valuer for a ~string enum type. It validates
// membership and returns a plain string, which is a valid driver.Value regardless
// of how a given driver's converter treats named types.
func ValueString[T ~string](set *Set[T], value T) (driver.Value, error) {
	if !set.Has(value) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidValue, string(value))
	}

	return string(value), nil
}

// MarshalJSONString implements json.Marshaler for a ~string enum type and validates
// membership before encoding the value as a JSON string.
func MarshalJSONString[T ~string](set *Set[T], value T) ([]byte, error) {
	if !set.Has(value) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidValue, string(value))
	}

	return json.Marshal(string(value))
}

// UnmarshalJSONString implements json.Unmarshaler for a ~string enum type. It decodes
// a JSON string and validates membership before storing the value into dst.
func UnmarshalJSONString[T ~string](set *Set[T], dst *T, data []byte) error {
	var text string

	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}

	return assignString(set, dst, text)
}
