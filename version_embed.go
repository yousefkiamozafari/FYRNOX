package vpnclient

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var embeddedVersion string

// Version is the authoritative release version embedded from the root VERSION file.
var Version = strings.TrimSpace(embeddedVersion)
