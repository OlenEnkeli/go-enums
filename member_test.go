package enum

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"math"
	"slices"
	"testing"
)

// color is an AutoEnum test type built on the Member wrapper alias.
type color = Member[string]

var (
	red   = Of("red")
	green = Of("green")
	blue  = Of("blue")

	colors = New(red, green, blue)
)

// Compile-time checks that Member satisfies the standard interfaces.
var (
	_ driver.Valuer    = Member[string]{}
	_ json.Marshaler   = Member[string]{}
	_ sql.Scanner      = (*Member[string])(nil)
	_ json.Unmarshaler = (*Member[string])(nil)
)

func TestMemberAccessors(t *testing.T) {
	member := Of("red")

	if got := member.Get(); got != "red" {
		t.Errorf("Get() = %q, want %q", got, "red")
	}

	if got := member.String(); got != "red" {
		t.Errorf("String() = %q, want %q", got, "red")
	}

	if got := Of(42).String(); got != "42" {
		t.Errorf("String() of int = %q, want %q", got, "42")
	}
}

func TestMemberJSONRoundTrip(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		data, err := json.Marshal(red)
		if err != nil {
			t.Fatalf("Marshal error = %v", err)
		}

		if got, want := string(data), `"red"`; got != want {
			t.Fatalf("Marshal = %s, want %s", got, want)
		}

		var got color

		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if got != red {
			t.Fatalf("round trip = %v, want %v", got, red)
		}
	})

	t.Run("int", func(t *testing.T) {
		member := Of(7)

		data, err := json.Marshal(member)
		if err != nil {
			t.Fatalf("Marshal error = %v", err)
		}

		if got, want := string(data), "7"; got != want {
			t.Fatalf("Marshal = %s, want %s", got, want)
		}

		var got Member[int]

		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal error = %v", err)
		}

		if got != member {
			t.Fatalf("round trip = %v, want %v", got, member)
		}
	})
}

func TestMemberValue(t *testing.T) {
	testCases := []struct {
		name    string
		value   func() (driver.Value, error)
		want    driver.Value
		wantErr error
	}{
		{"string", Of("red").Value, "red", nil},
		{"int", Of(5).Value, int64(5), nil},
		{"uint", Of(uint(5)).Value, int64(5), nil},
		{"float", Of(1.5).Value, 1.5, nil},
		{"bool", Of(true).Value, true, nil},
		{"uint64 overflow", Of(uint64(math.MaxUint64)).Value, nil, ErrUnsupportedType},
		{"unsupported kind", Of(struct{ X int }{1}).Value, nil, ErrUnsupportedType},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := testCase.value()
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("Value() err = %v, want %v", err, testCase.wantErr)
			}

			if testCase.wantErr == nil && got != testCase.want {
				t.Fatalf("Value() = %#v (%T), want %#v (%T)", got, got, testCase.want, testCase.want)
			}
		})
	}
}

func TestMemberScanString(t *testing.T) {
	testCases := []struct {
		name    string
		src     any
		want    color
		wantErr error
	}{
		{"from string", "blue", blue, nil},
		{"from bytes", []byte("green"), green, nil},
		{"nil is null", nil, color{}, ErrNullValue},
		{"unsupported type", 42, color{}, ErrUnsupportedType},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var got color

			err := got.Scan(testCase.src)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("Scan(%#v) err = %v, want %v", testCase.src, err, testCase.wantErr)
			}

			if testCase.wantErr == nil && got != testCase.want {
				t.Fatalf("Scan(%#v) = %v, want %v", testCase.src, got, testCase.want)
			}
		})
	}
}

func TestMemberScanNumeric(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		var member Member[int]

		if err := member.Scan(int64(9)); err != nil {
			t.Fatalf("Scan error = %v", err)
		}

		if member.Get() != 9 {
			t.Fatalf("Scan = %d, want 9", member.Get())
		}
	})

	t.Run("int overflow", func(t *testing.T) {
		var member Member[int8]

		if err := member.Scan(int64(1000)); !errors.Is(err, ErrUnsupportedType) {
			t.Fatalf("Scan overflow err = %v, want ErrUnsupportedType", err)
		}
	})

	t.Run("uint", func(t *testing.T) {
		var member Member[uint]

		if err := member.Scan(int64(5)); err != nil {
			t.Fatalf("Scan error = %v", err)
		}

		if member.Get() != 5 {
			t.Fatalf("Scan = %d, want 5", member.Get())
		}
	})

	t.Run("uint overflow", func(t *testing.T) {
		var member Member[uint8]

		if err := member.Scan(int64(1000)); !errors.Is(err, ErrUnsupportedType) {
			t.Fatalf("Scan overflow err = %v, want ErrUnsupportedType", err)
		}
	})

	t.Run("uint negative", func(t *testing.T) {
		var member Member[uint]

		if err := member.Scan(int64(-1)); !errors.Is(err, ErrUnsupportedType) {
			t.Fatalf("Scan negative err = %v, want ErrUnsupportedType", err)
		}
	})

	t.Run("bool", func(t *testing.T) {
		var member Member[bool]

		if err := member.Scan(true); err != nil {
			t.Fatalf("Scan error = %v", err)
		}

		if !member.Get() {
			t.Fatal("Scan = false, want true")
		}
	})

	t.Run("float from float64", func(t *testing.T) {
		var member Member[float64]

		if err := member.Scan(float64(1.5)); err != nil {
			t.Fatalf("Scan error = %v", err)
		}

		if member.Get() != 1.5 {
			t.Fatalf("Scan = %v, want 1.5", member.Get())
		}
	})

	t.Run("float from int64", func(t *testing.T) {
		var member Member[float64]

		if err := member.Scan(int64(3)); err != nil {
			t.Fatalf("Scan error = %v", err)
		}

		if member.Get() != 3 {
			t.Fatalf("Scan = %v, want 3", member.Get())
		}
	})

	// Wrong source types per kind must fall through to ErrUnsupportedType.
	t.Run("wrong source per kind", func(t *testing.T) {
		cases := []struct {
			name string
			scan func(any) error
			src  any
		}{
			{"int from string", func(src any) error { var member Member[int]; return member.Scan(src) }, "5"},
			{"uint from string", func(src any) error { var member Member[uint]; return member.Scan(src) }, "5"},
			{"bool from string", func(src any) error { var member Member[bool]; return member.Scan(src) }, "true"},
			{"float from string", func(src any) error { var member Member[float64]; return member.Scan(src) }, "1.5"},
		}

		for _, testCase := range cases {
			t.Run(testCase.name, func(t *testing.T) {
				if err := testCase.scan(testCase.src); !errors.Is(err, ErrUnsupportedType) {
					t.Fatalf("Scan(%#v) err = %v, want ErrUnsupportedType", testCase.src, err)
				}
			})
		}
	})
}

func TestMemberValueScanRoundTrip(t *testing.T) {
	for _, want := range colors.Members() {
		stored, err := want.Value()
		if err != nil {
			t.Fatalf("Value(%v) error = %v", want, err)
		}

		var got color

		if err := got.Scan(stored); err != nil {
			t.Fatalf("Scan(%#v) error = %v", stored, err)
		}

		if got != want {
			t.Fatalf("round trip = %v, want %v", got, want)
		}
	}
}

func TestEnumParse(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		want color
		ok   bool
	}{
		{"member", "red", red, true},
		{"missing", "pink", color{}, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, ok := colors.Parse(testCase.in)
			if ok != testCase.ok {
				t.Fatalf("Parse(%q) ok = %v, want %v", testCase.in, ok, testCase.ok)
			}

			if got != testCase.want {
				t.Fatalf("Parse(%q) = %v, want %v", testCase.in, got, testCase.want)
			}
		})
	}
}

func TestEnumContains(t *testing.T) {
	if !colors.Contains(red) {
		t.Error("Contains(red) = false, want true")
	}

	if colors.Contains(Of("pink")) {
		t.Error("Contains(pink) = true, want false")
	}
}

func TestEnumMembersAndLen(t *testing.T) {
	if got, want := colors.Len(), 3; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}

	if got, want := colors.Members(), []color{red, green, blue}; !slices.Equal(got, want) {
		t.Fatalf("Members() = %v, want %v", got, want)
	}
}

func TestEnumAll(t *testing.T) {
	var collected []color

	for member := range colors.All() {
		collected = append(collected, member)
	}

	if want := []color{red, green, blue}; !slices.Equal(collected, want) {
		t.Fatalf("All() yielded %v, want %v", collected, want)
	}
}

func TestEnumAllEarlyBreak(t *testing.T) {
	var collected []color

	for member := range colors.All() {
		collected = append(collected, member)

		break
	}

	if want := []color{red}; !slices.Equal(collected, want) {
		t.Fatalf("All() with break yielded %v, want %v", collected, want)
	}
}
