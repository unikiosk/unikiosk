//go:build tools
// +build tools

package main

// Tooling not managed by go mod and being used in build or development workflow

import (
	_ "filippo.io/mkcert"
	_ "golang.org/x/tools/cmd/goimports"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
