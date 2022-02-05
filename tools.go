//go:build tools

package do

// This file is only used to list dependencies of binaries we use for development.
// It uses a build tag to ensure the file is not picked up when building our app.
import (
	// goreadme is used to generate our README.md from godoc
	_ "github.com/posener/goreadme/cmd/goreadme"
)
