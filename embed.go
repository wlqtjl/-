package web

import "embed"

//go:embed *.html *.css *.js
var StaticFS embed.FS
