package enum

import (
	"slices"
	"testing"
)

func TestNewSetDeduplicatesPreservingOrder(t *testing.T) {
	set := NewSet("b", "a", "b", "c", "a")

	if got, want := set.Len(), 3; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}

	if got, want := set.Values(), []string{"b", "a", "c"}; !slices.Equal(got, want) {
		t.Fatalf("Values() = %v, want %v", got, want)
	}
}

func TestSetHas(t *testing.T) {
	set := NewSet(1, 2, 3)

	testCases := []struct {
		name string
		in   int
		want bool
	}{
		{"member", 2, true},
		{"first", 1, true},
		{"missing", 4, false},
		{"zero", 0, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := set.Has(testCase.in); got != testCase.want {
				t.Errorf("Has(%d) = %v, want %v", testCase.in, got, testCase.want)
			}
		})
	}
}

func TestSetValuesReturnsCopy(t *testing.T) {
	set := NewSet("a", "b")

	got := set.Values()
	got[0] = "mutated"

	if fresh := set.Values(); fresh[0] != "a" {
		t.Fatalf("Values() leaked internal slice: got %q after caller mutation", fresh[0])
	}
}

func TestSetAll(t *testing.T) {
	set := NewSet("x", "y", "z")

	var collected []string

	for value := range set.All() {
		collected = append(collected, value)
	}

	if want := []string{"x", "y", "z"}; !slices.Equal(collected, want) {
		t.Fatalf("All() yielded %v, want %v", collected, want)
	}
}

func TestSetAllEarlyBreak(t *testing.T) {
	set := NewSet("x", "y", "z")

	var collected []string

	for value := range set.All() {
		collected = append(collected, value)

		break
	}

	if want := []string{"x"}; !slices.Equal(collected, want) {
		t.Fatalf("All() with break yielded %v, want %v", collected, want)
	}
}

func TestSetString(t *testing.T) {
	if got, want := NewSet("a", "b", "c").String(), "Set[a, b, c]"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}

	if got, want := NewSet[int]().String(), "Set[]"; got != want {
		t.Errorf("empty String() = %q, want %q", got, want)
	}
}
