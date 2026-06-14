package enum

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"iter"
	"math"
	"reflect"
)

// Member is a generic enum member wrapping a single comparable value of type T.
//
// It implements sql.Scanner, driver.Valuer, json.Marshaler and json.Unmarshaler once
// for every enum, so member types need zero per-type code. By Go's type system these
// methods cannot reach the specific Enum container, so they do NOT validate membership.
// Validate explicitly via Enum.Parse / Enum.Contains.
type Member[T comparable] struct {
	value T
}

// Of constructs a Member holding value. T is normally inferred from value.
func Of[T comparable](value T) Member[T] {
	return Member[T]{value: value}
}

// Get returns the underlying value.
func (member Member[T]) Get() T {
	return member.value
}

// String returns the underlying value formatted with the %v verb.
func (member Member[T]) String() string {
	return fmt.Sprintf("%v", member.value)
}

// MarshalJSON encodes the underlying value as JSON. It is reflect-free: encoding/json
// handles T directly. Membership is not validated here.
func (member Member[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(member.value)
}

// UnmarshalJSON decodes JSON into the underlying value. It is reflect-free and does
// not validate membership.
func (member *Member[T]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &member.value)
}

// Value implements driver.Valuer via the reflect scalar bridge. Membership is not
// validated here.
func (member Member[T]) Value() (driver.Value, error) {
	return scalarToDriver(member.value)
}

// Scan implements sql.Scanner via the reflect scalar bridge. A nil source yields
// ErrNullValue. Membership is not validated here.
func (member *Member[T]) Scan(src any) error {
	if src == nil {
		return ErrNullValue
	}

	return scanScalar(reflect.ValueOf(&member.value).Elem(), src)
}

// scalarToDriver converts a comparable scalar to a valid driver.Value. This is the
// only place the AutoEnum front-end uses reflect; it is confined to the SQL bridge.
// Supported kinds map unambiguously to driver.Value: string, signed/unsigned integers,
// floats and bool.
func scalarToDriver(value any) (driver.Value, error) {
	reflected := reflect.ValueOf(value)

	switch reflected.Kind() {
	case reflect.String:
		return reflected.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflected.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		unsigned := reflected.Uint()

		if unsigned > math.MaxInt64 {
			return nil, fmt.Errorf("%w: uint value %d overflows int64", ErrUnsupportedType, unsigned)
		}

		return int64(unsigned), nil
	case reflect.Float32, reflect.Float64:
		return reflected.Float(), nil
	case reflect.Bool:
		return reflected.Bool(), nil
	default:
		return nil, fmt.Errorf("%w: %T", ErrUnsupportedType, value)
	}
}

// scanScalar assigns a driver source value into the settable destination target, whose
// kind is the underlying kind of T. Sources follow the driver.Value contract
// (int64, float64, bool, []byte, string).
func scanScalar(target reflect.Value, src any) error {
	switch target.Kind() {
	case reflect.String:
		switch source := src.(type) {
		case string:
			target.SetString(source)

			return nil
		case []byte:
			target.SetString(string(source))

			return nil
		}
	case reflect.Bool:
		if boolean, ok := src.(bool); ok {
			target.SetBool(boolean)

			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if number, ok := src.(int64); ok {
			return scanInt(target, number)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if number, ok := src.(int64); ok {
			return scanUint(target, number)
		}
	case reflect.Float32, reflect.Float64:
		switch number := src.(type) {
		case float64:
			target.SetFloat(number)

			return nil
		case int64:
			target.SetFloat(float64(number))

			return nil
		}
	}

	return fmt.Errorf("%w: cannot scan %T into %s", ErrUnsupportedType, src, target.Type())
}

// scanInt stores a signed integer source into target, guarding against overflow.
func scanInt(target reflect.Value, number int64) error {
	if target.OverflowInt(number) {
		return fmt.Errorf("%w: value %d overflows %s", ErrUnsupportedType, number, target.Type())
	}

	target.SetInt(number)

	return nil
}

// scanUint stores a signed integer source into an unsigned target, guarding against
// negative values and overflow.
func scanUint(target reflect.Value, number int64) error {
	if number < 0 {
		return fmt.Errorf("%w: negative value %d into %s", ErrUnsupportedType, number, target.Type())
	}

	unsigned := uint64(number)

	if target.OverflowUint(unsigned) {
		return fmt.Errorf("%w: value %d overflows %s", ErrUnsupportedType, number, target.Type())
	}

	target.SetUint(unsigned)

	return nil
}

// Enum is the registry (container) of allowed Members. It is the explicit-validation
// counterpart to the non-validating Member methods.
type Enum[T comparable] struct {
	set *Set[T]
}

// New builds an Enum from the given members, preserving first-seen order.
func New[T comparable](members ...Member[T]) *Enum[T] {
	values := make([]T, len(members))

	for index, member := range members {
		values[index] = member.value
	}

	return &Enum[T]{set: NewSet(values...)}
}

// Parse returns the registered Member for the raw value. The bool reports whether
// value is a member; on false the zero Member is returned.
func (enumeration *Enum[T]) Parse(value T) (Member[T], bool) {
	if enumeration.set.Has(value) {
		return Member[T]{value: value}, true
	}

	return Member[T]{}, false
}

// Contains reports whether member is registered.
func (enumeration *Enum[T]) Contains(member Member[T]) bool {
	return enumeration.set.Has(member.value)
}

// Members returns the registered members in insertion order. The result is a fresh
// copy.
func (enumeration *Enum[T]) Members() []Member[T] {
	values := enumeration.set.Values()
	result := make([]Member[T], len(values))

	for index, value := range values {
		result[index] = Member[T]{value: value}
	}

	return result
}

// Len returns the number of registered members.
func (enumeration *Enum[T]) Len() int {
	return enumeration.set.Len()
}

// All returns an iterator over the registered members in insertion order.
func (enumeration *Enum[T]) All() iter.Seq[Member[T]] {
	return func(yield func(Member[T]) bool) {
		for _, value := range enumeration.set.Values() {
			if !yield(Member[T]{value: value}) {
				return
			}
		}
	}
}
