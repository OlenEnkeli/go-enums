package enum

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
)

// This file holds the StrictEnum delegate helpers for signed integer enum types.
//
// Like the ~string helpers, they validate membership at the boundary in both
// directions and use no reflect. Signed integers map unambiguously onto the int64
// scalar of driver.Value and JSON numbers; conversion to int64 is always exact for
// every signed kind, so no overflow guard is needed on the way out.

// Signed is the constraint for the integer underlying kinds supported by the integer
// StrictEnum helpers.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// ScanInt implements sql.Scanner for a signed integer enum type and validates
// membership. It accepts an int64 source (the canonical driver representation) or a
// textual []byte/string; nil yields ErrNullValue and any other type yields
// ErrUnsupportedType.
func ScanInt[T Signed](set *Set[T], dst *T, src any) error {
	switch source := src.(type) {
	case int64:
		return assignInt(set, dst, source)
	case []byte:
		return assignIntText(set, dst, string(source))
	case string:
		return assignIntText(set, dst, source)
	case nil:
		return ErrNullValue
	default:
		return fmt.Errorf("%w: %T", ErrUnsupportedType, src)
	}
}

// assignInt validates a candidate against the set and, on success, stores it into dst.
// The round-trip check rejects candidates that do not fit the enum's narrower kind
// (e.g. 1000 into an int8 enum), so truncated values can never be silently accepted.
func assignInt[T Signed](set *Set[T], dst *T, candidate int64) error {
	value := T(candidate)

	if int64(value) != candidate {
		return fmt.Errorf("%w: %d is out of range for the enum type", ErrInvalidValue, candidate)
	}

	if !set.Has(value) {
		return fmt.Errorf("%w: %d", ErrInvalidValue, candidate)
	}

	*dst = value

	return nil
}

// assignIntText parses a base-10 integer from text and delegates to assignInt.
func assignIntText[T Signed](set *Set[T], dst *T, text string) error {
	candidate, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidValue, text)
	}

	return assignInt(set, dst, candidate)
}

// ValueInt implements driver.Valuer for a signed integer enum type. It validates
// membership and returns a plain int64, which is a valid driver.Value regardless of
// how a given driver's converter treats named types.
func ValueInt[T Signed](set *Set[T], value T) (driver.Value, error) {
	if !set.Has(value) {
		return nil, fmt.Errorf("%w: %d", ErrInvalidValue, int64(value))
	}

	return int64(value), nil
}

// MarshalJSONInt implements json.Marshaler for a signed integer enum type and
// validates membership before encoding the value as a JSON number. It marshals the
// base int64, never the enum-typed value, to avoid recursing into the delegate.
func MarshalJSONInt[T Signed](set *Set[T], value T) ([]byte, error) {
	if !set.Has(value) {
		return nil, fmt.Errorf("%w: %d", ErrInvalidValue, int64(value))
	}

	return json.Marshal(int64(value))
}

// UnmarshalJSONInt implements json.Unmarshaler for a signed integer enum type. It
// decodes a JSON number into an int64 and validates membership before storing it into
// dst. Non-integer JSON numbers (e.g. 2.5) are rejected by the decoder.
func UnmarshalJSONInt[T Signed](set *Set[T], dst *T, data []byte) error {
	var candidate int64

	if err := json.Unmarshal(data, &candidate); err != nil {
		return err
	}

	return assignInt(set, dst, candidate)
}
