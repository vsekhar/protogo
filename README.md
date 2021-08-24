# protogo
Protogo is an experiment to speed up development of distributed back-end applications by making software easy to express for developers and easy to distribute for runtime systems.

Protogo is a subset of the [Go programming language](https://golang.org). You should be fairly familiar with Go before working with Protogo.

## Writing a protogo module

Protogo modules start as Go packages. The `protogo` tool, however, requires that the types of all exported variables, function arguments and function return values must be a struct with `protogo` field tags. Protogo will then synthesize a protocol buffer for each of these types and generate a wrapper for the package that marshalls these types between running instances. The output of this process is a protogo module.

## How Protogo works

Protogo analyzes a Go package and walks the types of its exports to build an exported type list. A type is added to the exported type list if any of the following is true:

  1. The type appears in the declaration of an exported package-level variable
  1. The type appears in the declaration of an exported package-level function argument or return list
  1. TODO: (pending decision on methods) The type appears in an exported method argument or return list
  1. The type appears as a field in a struct that is itself on the exported type list

Exported types must be structs and each field in the struct must have a `protogo` struct tag with a unique (within the struct) integer id for each field.

Protogo uses the ids to synthesize a protocol buffer for each exported type. Protogo then synthesizes a main function that starts a server and marshalls protocol buffers into and out of the module. Compiling this server produces a protogo module.

Go packages intended for Protogo must include the `//go:build protogo` build tag.

## Imports

Imports in Protogo are handled with the same rules as in Go. This means Protogo modules can use the standard library and all existing Go packages.

When a Protogo module imports another protogo module, the import and all accesses to exported variables and functions are rewritten into RPCs using protocol buffers.

## Execution

Protogo modules are executed on a protogo runtime, which will eventually be distributed.

Each protogo module is instatiated within a service container

### Open questions

  * Could synthesize a gRPC definition for exported functions (what about methods?)
  * What to do about methods? Where does the data live?
  * How to handle imports of protogo? Package build tag?
