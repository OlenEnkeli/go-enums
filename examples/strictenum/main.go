// Command strictenum demonstrates the StrictEnum style of go-enums: a bare string
// type whose one-line delegate methods validate every value at the boundary.
package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	enum "github.com/OlenEnkeli/go-enums"
)

// OperationStatus is an async operation status modelled as a StrictEnum.
type OperationStatus string

// Allowed OperationStatus values.
const (
	StatusNew        OperationStatus = "NEW"
	StatusProcessing OperationStatus = "PROCESSING"
	StatusDone       OperationStatus = "DONE"
)

// operationStatusSet is the registry of allowed values, built once.
var operationStatusSet = enum.NewSet(StatusNew, StatusProcessing, StatusDone)

// IsValid reports whether the status is a registered member.
func (status OperationStatus) IsValid() bool {
	return operationStatusSet.Has(status)
}

// Scan implements sql.Scanner, validating the value at the boundary.
func (status *OperationStatus) Scan(src any) error {
	return enum.ScanString(operationStatusSet, status, src)
}

// Value implements driver.Valuer, validating the value at the boundary.
func (status OperationStatus) Value() (driver.Value, error) {
	return enum.ValueString(operationStatusSet, status)
}

// MarshalJSON implements json.Marshaler, validating the value at the boundary.
func (status OperationStatus) MarshalJSON() ([]byte, error) {
	return enum.MarshalJSONString(operationStatusSet, status)
}

// UnmarshalJSON implements json.Unmarshaler, validating the value at the boundary.
func (status *OperationStatus) UnmarshalJSON(data []byte) error {
	return enum.UnmarshalJSONString(operationStatusSet, status, data)
}

func main() {
	fmt.Println("== StrictEnum: allowed values ==")

	for _, status := range operationStatusSet.Values() {
		fmt.Printf("  - %s (valid=%t)\n", status, status.IsValid())
	}

	fmt.Println("\n== Scan validates at the boundary ==")

	var parsed OperationStatus

	if err := parsed.Scan("PROCESSING"); err == nil {
		fmt.Printf("  Scan(\"PROCESSING\") -> %s\n", parsed)
	}

	if err := parsed.Scan("BOGUS"); err != nil {
		fmt.Printf("  Scan(\"BOGUS\") rejected: %v\n", err)
	}

	fmt.Println("\n== JSON is native and validated ==")

	encoded, err := json.Marshal(StatusDone)
	if err != nil {
		fmt.Printf("  marshal error: %v\n", err)

		return
	}

	fmt.Printf("  Marshal(StatusDone) -> %s\n", encoded)

	var decoded OperationStatus

	if err := json.Unmarshal(encoded, &decoded); err != nil {
		fmt.Printf("  unmarshal error: %v\n", err)

		return
	}

	fmt.Printf("  Unmarshal(%s) -> %s\n", encoded, decoded)
}
