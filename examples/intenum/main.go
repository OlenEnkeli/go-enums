// Command intenum demonstrates the StrictEnum style over a signed integer type: a
// bare int whose one-line delegate methods validate every value at the boundary.
package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	enum "github.com/OlenEnkeli/go-enums"
)

// Priority is a task priority modelled as an integer StrictEnum.
type Priority int

// Allowed Priority values.
const (
	PriorityLow    Priority = 1
	PriorityMedium Priority = 2
	PriorityHigh   Priority = 3
)

// prioritySet is the registry of allowed values, built once.
var prioritySet = enum.NewSet(PriorityLow, PriorityMedium, PriorityHigh)

// IsValid reports whether the priority is a registered member.
func (level Priority) IsValid() bool {
	return prioritySet.Has(level)
}

// Scan implements sql.Scanner, validating the value at the boundary.
func (level *Priority) Scan(src any) error {
	return enum.ScanInt(prioritySet, level, src)
}

// Value implements driver.Valuer, validating the value at the boundary.
func (level Priority) Value() (driver.Value, error) {
	return enum.ValueInt(prioritySet, level)
}

// MarshalJSON implements json.Marshaler, validating the value at the boundary.
func (level Priority) MarshalJSON() ([]byte, error) {
	return enum.MarshalJSONInt(prioritySet, level)
}

// UnmarshalJSON implements json.Unmarshaler, validating the value at the boundary.
func (level *Priority) UnmarshalJSON(data []byte) error {
	return enum.UnmarshalJSONInt(prioritySet, level, data)
}

func main() {
	fmt.Println("== StrictEnum (int): allowed values ==")

	for _, level := range prioritySet.Values() {
		fmt.Printf("  - %d (valid=%t)\n", level, level.IsValid())
	}

	fmt.Println("\n== Scan validates at the boundary ==")

	var parsed Priority

	if err := parsed.Scan(int64(2)); err == nil {
		fmt.Printf("  Scan(2) -> %d\n", parsed)
	}

	if err := parsed.Scan(int64(9)); err != nil {
		fmt.Printf("  Scan(9) rejected: %v\n", err)
	}

	fmt.Println("\n== JSON is a native number and validated ==")

	encoded, err := json.Marshal(PriorityHigh)
	if err != nil {
		fmt.Printf("  marshal error: %v\n", err)

		return
	}

	fmt.Printf("  Marshal(PriorityHigh) -> %s\n", encoded)

	var decoded Priority

	if err := json.Unmarshal(encoded, &decoded); err != nil {
		fmt.Printf("  unmarshal error: %v\n", err)

		return
	}

	fmt.Printf("  Unmarshal(%s) -> %d\n", encoded, decoded)
}
