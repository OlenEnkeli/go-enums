package enum

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"
)

// priority is a StrictEnum test type with a signed-int underlying kind.
type priority int

const (
	priorityLow  priority = 1
	priorityHigh priority = 3
)

var prioritySet = NewSet(priorityLow, priorityHigh)

func (level priority) IsValid() bool {
	return prioritySet.Has(level)
}

func (level *priority) Scan(src any) error {
	return ScanInt(prioritySet, level, src)
}

func (level priority) Value() (driver.Value, error) {
	return ValueInt(prioritySet, level)
}

func (level priority) MarshalJSON() ([]byte, error) {
	return MarshalJSONInt(prioritySet, level)
}

func (level *priority) UnmarshalJSON(data []byte) error {
	return UnmarshalJSONInt(prioritySet, level, data)
}

// narrowCode exercises the round-trip range guard: 0 is a member, so a wider value
// that wraps to 0 (e.g. 256) must still be rejected, not silently accepted.
type narrowCode int8

const codeZero narrowCode = 0

var narrowCodeSet = NewSet(codeZero)

func (code *narrowCode) Scan(src any) error {
	return ScanInt(narrowCodeSet, code, src)
}

// Compile-time checks that the delegates satisfy the standard interfaces.
var (
	_ driver.Valuer    = priorityLow
	_ json.Marshaler   = priorityLow
	_ json.Unmarshaler = (*priority)(nil)
)

func TestScanInt(t *testing.T) {
	testCases := []struct {
		name    string
		src     any
		want    priority
		wantErr error
	}{
		{"from int64", int64(3), priorityHigh, nil},
		{"from bytes", []byte("1"), priorityLow, nil},
		{"from string", "3", priorityHigh, nil},
		{"nil is null", nil, 0, ErrNullValue},
		{"unknown value", int64(2), 0, ErrInvalidValue},
		{"unknown text", "2", 0, ErrInvalidValue},
		{"non-numeric text", "high", 0, ErrInvalidValue},
		{"unsupported type", 1.5, 0, ErrUnsupportedType},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var level priority

			err := level.Scan(testCase.src)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("Scan(%#v) err = %v, want %v", testCase.src, err, testCase.wantErr)
			}

			if testCase.wantErr == nil && level != testCase.want {
				t.Fatalf("Scan(%#v) = %d, want %d", testCase.src, level, testCase.want)
			}
		})
	}
}

func TestScanIntRangeGuard(t *testing.T) {
	var code narrowCode

	// 256 wraps to int8(0); without the guard this would be silently accepted as
	// codeZero. The round-trip check must reject it instead.
	if err := code.Scan(int64(256)); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("Scan(256) into int8 enum err = %v, want ErrInvalidValue", err)
	}

	// A genuine member still scans.
	if err := code.Scan(int64(0)); err != nil {
		t.Fatalf("Scan(0) error = %v", err)
	}

	if code != codeZero {
		t.Fatalf("Scan(0) = %d, want %d", code, codeZero)
	}
}

func TestValueInt(t *testing.T) {
	stored, err := priorityHigh.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	// driver.Value must be a plain int64, not the named type.
	if number, ok := stored.(int64); !ok || number != 3 {
		t.Fatalf("Value() = %#v (%T), want int64 %d", stored, stored, 3)
	}

	if _, err := priority(2).Value(); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("Value() of unknown err = %v, want ErrInvalidValue", err)
	}
}

func TestIntValueScanRoundTrip(t *testing.T) {
	for _, want := range prioritySet.Values() {
		stored, err := want.Value()
		if err != nil {
			t.Fatalf("Value(%d) error = %v", want, err)
		}

		var got priority

		if err := got.Scan(stored); err != nil {
			t.Fatalf("Scan(%#v) error = %v", stored, err)
		}

		if got != want {
			t.Fatalf("round trip = %d, want %d", got, want)
		}
	}
}

func TestIntJSON(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		data, err := json.Marshal(priorityHigh)
		if err != nil {
			t.Fatalf("Marshal error = %v", err)
		}

		if got, want := string(data), "3"; got != want {
			t.Fatalf("Marshal = %s, want %s", got, want)
		}
	})

	t.Run("marshal unknown rejected", func(t *testing.T) {
		if _, err := json.Marshal(priority(2)); !errors.Is(err, ErrInvalidValue) {
			t.Fatalf("Marshal unknown err = %v, want ErrInvalidValue", err)
		}
	})

	t.Run("unmarshal", func(t *testing.T) {
		var level priority

		if err := json.Unmarshal([]byte("1"), &level); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if level != priorityLow {
			t.Fatalf("Unmarshal = %d, want %d", level, priorityLow)
		}
	})

	t.Run("unmarshal unknown rejected", func(t *testing.T) {
		var level priority

		if err := json.Unmarshal([]byte("2"), &level); !errors.Is(err, ErrInvalidValue) {
			t.Fatalf("Unmarshal unknown err = %v, want ErrInvalidValue", err)
		}
	})

	t.Run("unmarshal non-integer rejected", func(t *testing.T) {
		var level priority

		if err := json.Unmarshal([]byte("2.5"), &level); err == nil {
			t.Fatal("Unmarshal of non-integer expected an error, got nil")
		}
	})

	t.Run("round trip", func(t *testing.T) {
		for _, want := range prioritySet.Values() {
			data, err := json.Marshal(want)
			if err != nil {
				t.Fatalf("Marshal(%d) error = %v", want, err)
			}

			var got priority

			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal(%s) error = %v", data, err)
			}

			if got != want {
				t.Fatalf("JSON round trip = %d, want %d", got, want)
			}
		}
	})
}

func TestIntIsValid(t *testing.T) {
	if !priorityHigh.IsValid() {
		t.Error("priorityHigh.IsValid() = false, want true")
	}

	if priority(2).IsValid() {
		t.Error("priority(2).IsValid() = true, want false")
	}
}
