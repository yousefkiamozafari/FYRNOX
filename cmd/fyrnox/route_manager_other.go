//go:build !windows

package main

import (
	"errors"
	"time"
)

var (
	ErrRouteConflict = errors.New("route already exists on a different interface")
	ErrRouteNotFound = errors.New("route not found in kernel routing table")
)

type RouteRecord struct {
	DestinationPrefix string
	Mask              string
	NextHop           string
	InterfaceIndex    int
	Metric            int
	Family            int
}

type RouteManager struct {
	tunIfIndex    int
	trackedRoutes []RouteRecord
}

func NewRouteManager(tunIfIndex int) *RouteManager {
	return &RouteManager{tunIfIndex: tunIfIndex}
}

func (rm *RouteManager) CheckConflict(prefix, mask string, family int) (bool, error) {
	return false, nil
}

func (rm *RouteManager) AddRouteIPv4(prefix, mask, nextHop string, metric int) error {
	return nil
}

func (rm *RouteManager) AddRouteIPv6(prefix string, metric int) error {
	return nil
}

func (rm *RouteManager) VerifyRouteExists(prefix string, family int) (bool, error) {
	return true, nil
}

func (rm *RouteManager) VerifyBestRoute(remoteIP string, expectedIfIndex int) (bool, error) {
	return true, nil
}

func (rm *RouteManager) DetectMoreSpecificPhysicalRoutes() ([]string, error) {
	return nil, nil
}

func (rm *RouteManager) Rollback() error {
	rm.trackedRoutes = nil
	return nil
}

func (rm *RouteManager) TrackedCount() int {
	return len(rm.trackedRoutes)
}

func (rm *RouteManager) IsTracked(prefix string) bool {
	return false
}
