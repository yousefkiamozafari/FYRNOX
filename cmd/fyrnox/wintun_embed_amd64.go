//go:build windows && amd64

package main

import _ "embed"

//go:embed wintun_amd64.dll
var embeddedWintunDLL []byte
