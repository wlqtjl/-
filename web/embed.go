package web

import "embed"

//go:embed *.html *.css *.js *.json
var StaticFS embed.FS
