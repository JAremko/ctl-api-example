package main

import (
	"sync"            // Package to handle synchronization
	"github.com/JAremko/ctl-api-example/thermalcamera" // Importing Protobuf definitions
)

// DefaultState defines the default state of the thermal camera
type DefaultState struct {
	sync.Mutex   // Mutex for synchronization
	ZoomLevel    int32
	ColorScheme  thermalcamera.ColorScheme
	BatteryLevel int32
}

// GetZoomLevel is a thread-safe getter for the zoom level
func (ds *DefaultState) GetZoomLevel() int32 {
	ds.Lock()
	defer ds.Unlock()
	return ds.ZoomLevel
}

// GetColorScheme is a thread-safe getter for the color scheme
func (ds *DefaultState) GetColorScheme() thermalcamera.ColorScheme {
	ds.Lock()
	defer ds.Unlock()
	return ds.ColorScheme
}

// GetBatteryLevel is a thread-safe getter for the battery level
func (ds *DefaultState) GetBatteryLevel() int32 {
	ds.Lock()
	defer ds.Unlock()
	return ds.BatteryLevel
}

// UpdateZoomLevel is a thread-safe setter for the zoom level
func (ds *DefaultState) UpdateZoomLevel(level int32) {
	ds.Lock()
	defer ds.Unlock()
	ds.ZoomLevel = level
}

// UpdateColorScheme is a thread-safe setter for the color scheme
func (ds *DefaultState) UpdateColorScheme(scheme thermalcamera.ColorScheme) {
	ds.Lock()
	defer ds.Unlock()
	ds.ColorScheme = scheme
}

// UpdateBatteryLevel is a thread-safe setter for the battery level
func (ds *DefaultState) UpdateBatteryLevel(level int32) {
	ds.Lock()
	defer ds.Unlock()
	ds.BatteryLevel = level
}
