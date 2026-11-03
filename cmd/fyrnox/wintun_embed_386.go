//go:build windows && 386

package main

import _ "embed"

//go:embed wintun_386.dll
var embeddedWintunDLL []byte
