// package embedded contains Embedded assets, typically used for views,
// to help make our binary as portable as possible.
package embedded

import "embed"

//go:embed templates/*
var Templates embed.FS

//go:embed static/*
var StaticAssets embed.FS

//go:embed static/favicon.ico
var Favicon []byte
