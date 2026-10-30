package vpnclient

import "net/http"

const mozillaVPNUserAgent = fastlySolverUserAgent

func applyMozillaVPNHeaders(req *http.Request) {
	req.Header.Set("User-Agent", fastlySolverUserAgent)
	req.Header.Set("Accept", "application/json")
}
