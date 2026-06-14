package enum

import "errors"

var (
	// ErrInvalidValue is returned when a value is not a registered member of an enum.
	ErrInvalidValue = errors.New("enum: invalid value")
	// ErrNullValue is returned when a NULL/nil source is scanned into a non-pointer enum.
	ErrNullValue = errors.New("enum: null value")
	// ErrUnsupportedType is returned when a source type cannot be converted to the enum value.
	ErrUnsupportedType = errors.New("enum: unsupported type")
)
