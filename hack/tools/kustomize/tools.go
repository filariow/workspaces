//go:build tools
// +build tools

// This package imports things required by build scripts, to force `go mod` to see them as dependencies
package tools

import (
    _ "sigs.k8s.io/kustomize/kustomize/v5"
)
