//go:build tools

package main

// Blank import pins the templ CLI in go.mod so `go generate` can use it
// without a global install.
//
// Usage:
//
//	go generate ./...
import _ "github.com/a-h/templ/cmd/templ"

//go:generate go run github.com/a-h/templ/cmd/templ generate
