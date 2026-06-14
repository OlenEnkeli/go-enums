# 🌊 Go-enums

> Universal, generics-based enum library for Go.

**Go has no built-in enums**. 

The usual workaround — `type Foo string` + a `const (...)` block +  hand-written `switch`/validation boilerplate — is repetitive and easy to get wrong. 

So `go-enums` turns that pattern into a small, reusable **registry** built once via a constructor (a `map`, never`reflect`), and offers two ergonomic styles on top of it.

- **Module:** `github.com/OlenEnkeli/go-enums` (package `enum`)
- **Requires:** `Go 1.24+`

## 🎯️ Install

```sh
go get github.com/OlenEnkeli/go-enums
```

```go
import enum "github.com/OlenEnkeli/go-enums"
```

## 🍏️ Two styles, one core

| Style              | Per-type code              | Boundary validation                        | Underlying types        |
|--------------------|----------------------------|--------------------------------------------| ----------------------- |
| 🏔️ **StrictEnum** | ✍🏻  ~5 one-line methods   | ✅ **yes** — `Scan`/`Value`/JSON reject     | `~string`, signed `~int…~int64` |
| 🦄 **AutoEnum**    | ✅  none (zero boilerplate) | 🚫 **no** — validate explicitly via `Enum` | any `comparable`        |

Pick **StrictEnum** when you want the **type to police itself** at the SQL/JSON boundary. 

Pick **AutoEnum** when you want zero per-type code and are happy to **validate explicitly**.

## 🏔️ StrictEnum — validating delegates

The enum type stays a bare `~string` (or signed integer), so JSON and SQL are native. 

Each method is a one-line delegate to a reflect-free helper that validates membership in **both** directions: 
 - Unknown values are rejected on the **way in** (`Scan`, `UnmarshalJSON`) 
 - And on the **way out** (`Value`,`MarshalJSON`).

### ✍🏻 String based *[recommended]*

```go
type OperationStatus string

const (
	StatusNew        OperationStatus = "NEW"
	StatusProcessing OperationStatus = "PROCESSING"
	StatusDone       OperationStatus = "DONE"
)

// The registry of allowed values, built once.
var operationStatusSet = enum.NewSet(StatusNew, StatusProcessing, StatusDone)

func (status OperationStatus) IsValid() bool {
	return operationStatusSet.Has(status)
}

func (status *OperationStatus) Scan(src any) error {
	return enum.ScanString(operationStatusSet, status, src)
}

func (status OperationStatus) Value() (driver.Value, error) {
	return enum.ValueString(operationStatusSet, status)
}

func (status OperationStatus) MarshalJSON() ([]byte, error) {
	return enum.MarshalJSONString(operationStatusSet, status)
}

func (status *OperationStatus) UnmarshalJSON(data []byte) error {
	return enum.UnmarshalJSONString(operationStatusSet, status, data)
}
```

```go
var status OperationStatus

_ = status.Scan("PROCESSING")          // status == StatusProcessing
err := status.Scan("BOGUS")            // err wraps enum.ErrInvalidValue — rejected
ok := StatusDone.IsValid()             // true

for _, value := range operationStatusSet.Values() { /* ... */ }
```

### 🔟 Integer based

For signed integer enums (`~int … ~int64`, constrained by `enum.Signed`), use the `Int` helpers. 

They map onto the `int64` scalar of `driver.Value` and JSON numbers, and **reject values** that do not fit the
type or are not members.

```go
type Priority int

const (
	PriorityLow    Priority = 1
	PriorityMedium Priority = 2
	PriorityHigh   Priority = 3
)

var prioritySet = enum.NewSet(PriorityLow, PriorityMedium, PriorityHigh)

func (level Priority) IsValid() bool                  { return prioritySet.Has(level) }
func (level *Priority) Scan(src any) error            { return enum.ScanInt(prioritySet, level, src) }
func (level Priority) Value() (driver.Value, error)   { return enum.ValueInt(prioritySet, level) }
func (level Priority) MarshalJSON() ([]byte, error)   { return enum.MarshalJSONInt(prioritySet, level) }
func (level *Priority) UnmarshalJSON(b []byte) error  { return enum.UnmarshalJSONInt(prioritySet, level, b) }
```

## 🦄 AutoEnum — zero boilerplate

`Member[T]` wraps a value and implements `sql.Scanner`, `driver.Valuer`, `json.Marshaler` and `json.Unmarshaler` once for every enum. 

These methods are native but **do not** validate membership — by Go's type system a method on a generic `Member[T]` cannot reach the specific container.

Validate explicitly through the `Enum[T]` registry (`Parse` / `Contains`).

```go
// The alias keeps Member's methods.
type Color = enum.Member[string]

var (
	Red   = enum.Of("red")
	Green = enum.Of("green")
	Blue  = enum.Of("blue")

	// The registry (container) of allowed members.
	Colors = enum.New(Red, Green, Blue)
)
```

```go
parsed, ok := Colors.Parse("red")   // parsed == Red, ok == true
_, ok = Colors.Parse("pink")        // ok == false
Colors.Contains(Red)                // true
Red.Get()                           // "red"

for member := range Colors.All() { /* iter.Seq[Member[string]] */ }

// Native in SQL/JSON, but Scan/Marshal here do NOT validate membership:
type Row struct{ Favorite Color }   // stored as a plain varchar
data, _ := json.Marshal(Red)        // "red"
```

> **Why no self-validating `Member`?** 
> 
> Boundary self-validation and "one method set for all enums" are mutually exclusive in Go: 
> 
> Two enums sharing an underlying type are the same `Member[string]` and cannot be told apart
> 
> A distinct named type does not inherit `Member`'s methods. 
> 
> So AutoEnum trades boundary validation for zero boilerplate; 
> 
> StrictEnum makes the opposite trade.

## ✨ API at a glance

- **`Set[T comparable]`** (core registry): `NewSet`, `Has`, `Values`, `Len`, `All() iter.Seq[T]`,
  `String`.
- **String helpers** (`~string`): `ScanString`, `ValueString`, `MarshalJSONString`,
  `UnmarshalJSONString`.
- **Integer helpers** (`Signed` = `~int … ~int64`): `ScanInt`, `ValueInt`, `MarshalJSONInt`,
  `UnmarshalJSONInt`.
- **`Member[T comparable]`**: `Of`, `Get`, `String`, plus native `Scan` / `Value` / `MarshalJSON` /
  `UnmarshalJSON`.
- **`Enum[T comparable]`**: `New`, `Parse`, `Contains`, `Members`, `Len`, `All() iter.Seq[Member[T]]`.
- **Errors:** `ErrInvalidValue`, `ErrNullValue`, `ErrUnsupportedType` (all wrapped; match with
  `errors.Is`).

`Scan` accepts:
 - The natural driver sources (`string`/`[]byte` for string enums; `int64` or textual `[]byte`/`string` for integer enums); 
 - A `nil` source yields `ErrNullValue` 
 - Anything else `ErrUnsupportedType`.

## 🤔 Examples

Runnable programs live under [`examples/`](examples):

```sh
go run ./examples/strictenum   # StrictEnum over a string
go run ./examples/intenum      # StrictEnum over a signed int
go run ./examples/autoenum     # AutoEnum via the Member wrapper
```

## 🛠️ Development

The workflow is driven by [Task](https://taskfile.dev):

**The discussions and PRs are highly welcome!**

```sh
task init          # install pinned golangci-lint and tidy modules (first run)
task lint          # standard linting (format check, go vet) + golangci-lint
task test          # go test -race -cover ./...
task check         # full gate: lint + test + build
task run-examples  # run all example programs
```

## ™️  License

See [LICENSE](LICENSE).
