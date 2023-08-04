# Table of Contents
- [Introduction](#introduction)
- [DefaultState Structure](#defaultstate-structure)
- [Thread-Safe Getters](#thread-safe-getters)
  - [GetZoomLevel](#getzoomlevel)
  - [GetColorScheme](#getcolorscheme)
  - [GetBatteryLevel](#getbatterylevel)
- [Thread-Safe Setters](#thread-safe-setters)
  - [UpdateZoomLevel](#updatezoomlevel)
  - [UpdateColorScheme](#updatecolorscheme)
  - [UpdateBatteryLevel](#updatebatterylevel)

# Introduction
The `state.go` file defines the default state of the thermal camera and provides thread-safe getters and setters for the zoom level, color scheme, and battery level. It uses the `sync.Mutex` package to ensure that the state can be accessed and modified by multiple threads without conflicts.

# DefaultState Structure
The `DefaultState` structure defines the default state of the thermal camera, including the zoom level, color scheme, and battery level.

```go
type DefaultState struct {
	sync.Mutex   // Mutex for synchronization
	ZoomLevel    int32
	ColorScheme  thermalcamera.ColorScheme
	BatteryLevel int32
}
```

# Thread-Safe Getters
These functions provide safe access to the state's properties.

## GetZoomLevel
Returns the current zoom level of the camera.

```go
func (ds *DefaultState) GetZoomLevel() int32 {
	ds.Lock()
	defer ds.Unlock()
	return ds.ZoomLevel
}
```

## GetColorScheme
Returns the current color scheme of the camera.

```go
func (ds *DefaultState) GetColorScheme() thermalcamera.ColorScheme {
	ds.Lock()
	defer ds.Unlock()
	return ds.ColorScheme
}
```

## GetBatteryLevel
Returns the current battery level of the camera.

```go
func (ds *DefaultState) GetBatteryLevel() int32 {
	ds.Lock()
	defer ds.Unlock()
	return ds.BatteryLevel
}
```

# Thread-Safe Setters
These functions provide safe modification of the state's properties.

## UpdateZoomLevel
Updates the zoom level of the camera.

```go
func (ds *DefaultState) UpdateZoomLevel(level int32) {
	ds.Lock()
	defer ds.Unlock()
	ds.ZoomLevel = level
}
```

## UpdateColorScheme
Updates the color scheme of the camera.

```go
func (ds *DefaultState) UpdateColorScheme(scheme thermalcamera.ColorScheme) {
	ds.Lock()
	defer ds.Unlock()
	ds.ColorScheme = scheme
}
```

## UpdateBatteryLevel
Updates the battery level of the camera.

```go
func (ds *DefaultState) UpdateBatteryLevel(level int32) {
	ds.Lock()
	defer ds.Unlock()
	ds.BatteryLevel = level
}
```
