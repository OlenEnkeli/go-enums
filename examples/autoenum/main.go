// Command autoenum demonstrates the AutoEnum style of go-enums: zero per-type code
// via the Member wrapper, with explicit validation through the Enum container.
package main

import (
	"encoding/json"
	"fmt"

	enum "github.com/OlenEnkeli/go-enums"
)

// Color is an AutoEnum built on the Member wrapper. The alias keeps Member's methods.
type Color = enum.Member[string]

// Color members and their registry.
var (
	Red   = enum.Of("red")
	Green = enum.Of("green")
	Blue  = enum.Of("blue")

	// Colors is the registry (container) of allowed members.
	Colors = enum.New(Red, Green, Blue)
)

func main() {
	fmt.Println("== AutoEnum: registered members ==")

	for member := range Colors.All() {
		fmt.Printf("  - %s\n", member)
	}

	fmt.Println("\n== Parse validates explicitly ==")

	if parsed, ok := Colors.Parse("red"); ok {
		fmt.Printf("  Parse(\"red\") -> %s\n", parsed.Get())
	}

	if _, ok := Colors.Parse("pink"); !ok {
		fmt.Println("  Parse(\"pink\") -> not a member")
	}

	fmt.Printf("  Contains(Red) = %t\n", Colors.Contains(Red))

	fmt.Println("\n== JSON is native; Scan/Marshal do NOT validate membership ==")

	encoded, err := json.Marshal(Red)
	if err != nil {
		fmt.Printf("  marshal error: %v\n", err)

		return
	}

	fmt.Printf("  Marshal(Red) -> %s\n", encoded)

	var decoded Color

	if err := json.Unmarshal([]byte(`"green"`), &decoded); err != nil {
		fmt.Printf("  unmarshal error: %v\n", err)

		return
	}

	fmt.Printf("  Unmarshal(\"green\") -> %s; Contains -> %t\n", decoded.Get(), Colors.Contains(decoded))
}
