package main

import (
	"strings"
	"sync"
)

var defaultGameSignatures = []string{
	"cs2.exe",
	"valorant-win64-shipping.exe",
	"dota2.exe",
	"leagueclient.exe",
	"r5apex.exe",
	"cod.exe",
	"overwatch.exe",
	"rainbowsix.exe",
	"fortniteclient-win64-shipping.exe",
	"pubg.exe",
	"rocketleague.exe",
	"gta5.exe",
}

type PingShieldManager struct {
	mu           sync.RWMutex
	customGames  []string
	detectedGame string
	gameDetected bool
}

var globalPingShield = &PingShieldManager{
	customGames: make([]string, 0),
}

func (p *PingShieldManager) SetCustomGames(games []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.customGames = games
}

func (p *PingShieldManager) GetCustomGames() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	res := make([]string, len(p.customGames))
	copy(res, p.customGames)
	return res
}

func (p *PingShieldManager) ScanActiveGameSockets() (bool, string) {
	// Refresh Windows extended TCP & UDP tables
	globalPortCache.refresh(false)

	p.mu.Lock()
	defer p.mu.Unlock()

	signatures := make(map[string]struct{}, len(defaultGameSignatures)+len(p.customGames))
	for _, g := range defaultGameSignatures {
		signatures[strings.ToLower(strings.TrimSpace(g))] = struct{}{}
	}
	for _, g := range p.customGames {
		clean := strings.ToLower(strings.TrimSpace(g))
		if clean != "" {
			signatures[clean] = struct{}{}
		}
	}

	globalPortCache.mu.RLock()
	defer globalPortCache.mu.RUnlock()

	// Check UDP sockets first (most game netcode uses UDP)
	for _, exe := range globalPortCache.udpPorts {
		lowExe := strings.ToLower(exe)
		if _, exists := signatures[lowExe]; exists {
			p.gameDetected = true
			p.detectedGame = exe
			return true, exe
		}
	}

	// Check TCP sockets
	for _, exe := range globalPortCache.tcpPorts {
		lowExe := strings.ToLower(exe)
		if _, exists := signatures[lowExe]; exists {
			p.gameDetected = true
			p.detectedGame = exe
			return true, exe
		}
	}

	p.gameDetected = false
	p.detectedGame = ""
	return false, ""
}

func (p *PingShieldManager) IsGameDetected() (bool, string) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.gameDetected, p.detectedGame
}

func (p *PingShieldManager) AddCustomGameSignature(exe string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	clean := strings.ToLower(strings.TrimSpace(exe))
	if clean == "" {
		return
	}
	for _, g := range p.customGames {
		if strings.ToLower(g) == clean {
			return
		}
	}
	p.customGames = append(p.customGames, clean)
}

func (p *PingShieldManager) RemoveCustomGameSignature(exe string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	clean := strings.ToLower(strings.TrimSpace(exe))
	filtered := make([]string, 0, len(p.customGames))
	for _, g := range p.customGames {
		if strings.ToLower(g) != clean {
			filtered = append(filtered, g)
		}
	}
	p.customGames = filtered
}

func (p *PingShieldManager) GetGameSignatures() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	all := make([]string, 0, len(defaultGameSignatures)+len(p.customGames))
	all = append(all, defaultGameSignatures...)
	all = append(all, p.customGames...)
	return all
}
