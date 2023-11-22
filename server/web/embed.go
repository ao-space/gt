//go:build release || !nostatic

package web

import "embed"

//go:embed dist/*
var embeddedFS embed.FS

func init() {
	FS = embeddedFS
}
