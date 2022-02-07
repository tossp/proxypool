package config

import "embed"

//go:embed assets/html/*
var HtmlFs embed.FS

//go:embed assets/static/*
var StaticFS embed.FS

//go:embed assets/flags.json
var GeoIpFS embed.FS
