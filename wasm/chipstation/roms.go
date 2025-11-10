package main

import (
	"embed"
)

//go:embed roms
var romAssets embed.FS

//go:embed assets
var assets embed.FS
