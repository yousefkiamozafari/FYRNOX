package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func defaultSettings() VpnSettings {
	return VpnSettings{
		SocksPort:          20808,
		HttpPort:           20809,
		UseSystemProxy:     true,
		EnableTunMode:      false,
		UseH3:              false,
		BypassRules:        "<local>;localhost;127.*;10.*;172.16.*;192.168.*",
		DirectIranTraffic:  true,
		SplitTunnelEnabled: false,
		SplitTunnelMode:    "bypass",
		SplitTunnelApps:    nil,
		GamingMode:         false,
		IPProtocol:         "dual", // "dual", "ipv4", "ipv6"
		KillSwitch:         false,
		AutoReconnect:      true,
	}
}

func getSettingsFilePath() string {
	return filepath.Join(getAppDataDir(), "settings.json")
}

func loadSettings() VpnSettings {
	s := defaultSettings()
	data, err := os.ReadFile(getSettingsFilePath())
	if err != nil {
		return s
	}

	var loaded VpnSettings
	if err := json.Unmarshal(data, &loaded); err != nil {
		return s
	}

	// Apply loaded values with fallback validation
	if loaded.SocksPort <= 0 {
		loaded.SocksPort = 20808
	}
	if loaded.HttpPort <= 0 {
		loaded.HttpPort = 20809
	}
	if loaded.IPProtocol == "" {
		loaded.IPProtocol = "dual"
	}
	return loaded
}

func saveSettings(s VpnSettings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getSettingsFilePath(), data, 0644)
}

func loadSettingsFile(path string) (VpnSettings, error) {
	s := defaultSettings()
	data, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(data, &s)
	return s, err
}

func saveSettingsFile(path string, s VpnSettings) ([]byte, error) {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, os.WriteFile(path, data, 0644)
}
