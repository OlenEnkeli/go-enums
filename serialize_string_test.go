package enum

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"
)

// operationStatus is a StrictEnum test type: a bare string with one-line delegates.
type operationStatus string

const (
	statusNew  operationStatus = "NEW"
	statusDone operationStatus = "DONE"
)

var operationStatusSet = NewSet(statusNew, statusDone)

func (status operationStatus) IsValid() bool {
	return operationStatusSet.Has(status)
}

func (status *operationStatus) Scan(src any) error {
	return ScanString(operationStatusSet, status, src)
}

func (status operationStatus) Value() (driver.Value, error) {
	return ValueString(operationStatusSet, status)
}

func (status operationStatus) MarshalJSON() ([]byte, error) {
	return MarshalJSONString(operationStatusSet, status)
}

func (status *operationStatus) UnmarshalJSON(data []byte) error {
	return UnmarshalJSONString(operationStatusSet, status, data)
}

// Compile-time checks that the delegates satisfy the standard interfaces.
var (
	_ driver.Valuer    = operationStatus("")
	_ json.Marshaler   = operationStatus("")
	_ json.Unmarshaler = (*operationStatus)(nil)
)

func TestScanString(t *testing.T) {
	testCases := []struct {
		name    string
		src     any
		want    operationStatus
		wantErr error
	}{
		{"from string", "DONE", statusDone, nil},
		{"from bytes", []byte("NEW"), statusNew, nil},
		{"nil is null", nil, "", ErrNullValue},
		{"unknown string", "BOGUS", "", ErrInvalidValue},
		{"unknown bytes", []byte("BOGUS"), "", ErrInvalidValue},
		{"unsupported type", 42, "", ErrUnsupportedType},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var status operationStatus

			err := status.Scan(testCase.src)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("Scan(%#v) err = %v, want %v", testCase.src, err, testCase.wantErr)
			}

			if testCase.wantErr == nil && status != testCase.want {
				t.Fatalf("Scan(%#v) = %q, want %q", testCase.src, status, testCase.want)
			}
		})
	}
}

func TestValueString(t *testing.T) {
	stored, err := statusDone.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	// driver.Value must be a plain string, not the named type.
	if text, ok := stored.(string); !ok || text != "DONE" {
		t.Fatalf("Value() = %#v (%T), want string %q", stored, stored, "DONE")
	}

	if _, err := operationStatus("BOGUS").Value(); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("Value() of unknown err = %v, want ErrInvalidValue", err)
	}
}

func TestStringValueScanRoundTrip(t *testing.T) {
	for _, want := range operationStatusSet.Values() {
		stored, err := want.Value()
		if err != nil {
			t.Fatalf("Value(%q) error = %v", want, err)
		}

		var got operationStatus

		if err := got.Scan(stored); err != nil {
			t.Fatalf("Scan(%#v) error = %v", stored, err)
		}

		if got != want {
			t.Fatalf("round trip = %q, want %q", got, want)
		}
	}
}

func TestStringJSON(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		data, err := json.Marshal(statusNew)
		if err != nil {
			t.Fatalf("Marshal error = %v", err)
		}

		if got, want := string(data), `"NEW"`; got != want {
			t.Fatalf("Marshal = %s, want %s", got, want)
		}
	})

	t.Run("marshal unknown rejected", func(t *testing.T) {
		if _, err := json.Marshal(operationStatus("BOGUS")); !errors.Is(err, ErrInvalidValue) {
			t.Fatalf("Marshal unknown err = %v, want ErrInvalidValue", err)
		}
	})

	t.Run("unmarshal", func(t *testing.T) {
		var status operationStatus

		if err := json.Unmarshal([]byte(`"DONE"`), &status); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if status != statusDone {
			t.Fatalf("Unmarshal = %q, want %q", status, statusDone)
		}
	})

	t.Run("unmarshal unknown rejected", func(t *testing.T) {
		var status operationStatus

		if err := json.Unmarshal([]byte(`"BOGUS"`), &status); !errors.Is(err, ErrInvalidValue) {
			t.Fatalf("Unmarshal unknown err = %v, want ErrInvalidValue", err)
		}
	})

	t.Run("unmarshal invalid json", func(t *testing.T) {
		var status operationStatus

		if err := json.Unmarshal([]byte(`123`), &status); err == nil {
			t.Fatal("Unmarshal of non-string expected an error, got nil")
		}
	})

	t.Run("round trip", func(t *testing.T) {
		for _, want := range operationStatusSet.Values() {
			data, err := json.Marshal(want)
			if err != nil {
				t.Fatalf("Marshal(%q) error = %v", want, err)
			}

			var got operationStatus

			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal(%s) error = %v", data, err)
			}

			if got != want {
				t.Fatalf("JSON round trip = %q, want %q", got, want)
			}
		}
	})
}

func TestStringIsValid(t *testing.T) {
	if !statusNew.IsValid() {
		t.Error("statusNew.IsValid() = false, want true")
	}

	if operationStatus("BOGUS").IsValid() {
		t.Error("BOGUS.IsValid() = true, want false")
	}
}
