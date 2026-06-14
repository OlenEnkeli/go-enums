package enum

import (
	"fmt"
	"iter"
	"strings"
)

// Set is a reflect-free registry of allowed enum values for any comparable type T.
//
// It is the core shared by both enum styles (StrictEnum and AutoEnum): a membership
// check plus deterministic enumeration. The registry is built once via NewSet and is
// read-only afterwards, so it is safe for concurrent reads.
type Set[T comparable] struct {
	members map[T]struct{}
	values  []T // insertion order, de-duplicated; kept for deterministic enumeration
}

// NewSet builds a Set from the given values. Duplicates are stored once and the
// first-seen insertion order is preserved for Values, All and String.
func NewSet[T comparable](values ...T) *Set[T] {
	set := &Set[T]{
		members: make(map[T]struct{}, len(values)),
		values:  make([]T, 0, len(values)),
	}

	for _, value := range values {
		if _, exists := set.members[value]; exists {
			continue
		}

		set.members[value] = struct{}{}
		set.values = append(set.values, value)
	}

	return set
}

// Has reports whether value is a registered member of the Set.
func (set *Set[T]) Has(value T) bool {
	_, exists := set.members[value]

	return exists
}

// Values returns the registered values in insertion order. The result is a fresh
// copy, so callers may modify it without affecting the Set.
func (set *Set[T]) Values() []T {
	result := make([]T, len(set.values))
	copy(result, set.values)

	return result
}

// Len returns the number of registered values.
func (set *Set[T]) Len() int {
	return len(set.values)
}

// All returns an iterator over the registered values in insertion order.
func (set *Set[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, value := range set.values {
			if !yield(value) {
				return
			}
		}
	}
}

// String returns a human-readable representation of the Set, e.g. "Set[a, b, c]".
func (set *Set[T]) String() string {
	parts := make([]string, len(set.values))

	for index, value := range set.values {
		parts[index] = fmt.Sprintf("%v", value)
	}

	return "Set[" + strings.Join(parts, ", ") + "]"
}
