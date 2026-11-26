package main

import vpnclient "firefox-vpn-client"

// AppVersion resolves from the embedded root VERSION file unless overridden at build time via -ldflags "-X main.AppVersion=...".
var AppVersion = vpnclient.Version
