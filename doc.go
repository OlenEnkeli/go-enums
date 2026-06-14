// Package enum turns the usual `type Foo string` + const + hand-written switch
// boilerplate into a reusable, generics-based registry.
//
// It offers two styles over one shared core:
//
//   - StrictEnum: a bare ~string type with one-line delegate methods that
//     validate at the boundary (ScanString, ValueString, MarshalJSONString,
//     UnmarshalJSONString).
//   - AutoEnum: the Member[T] wrapper with zero per-type code and explicit
//     validation via the Enum[T] container.
//
// The set of allowed values is a registry (Set[T]) built once via a constructor,
// never via reflect.
package enum
