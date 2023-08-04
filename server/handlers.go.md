# Table of Contents
- [Introduction](#introduction)
- [Data from WebSockets](#data-from-websockets)
  - [HandleSetZoomLevel](#handlesetzoomlevel)
  - [HandleSetColorScheme](#handlecolorscheme)
- [Data from C Pipes](#data-from-c-pipes)
  - [HandlePacketsFromC](#handlepacketsfromc)
- [Channel Data Flow](#channel-data-flow)

# Introduction
The `handlers.go` file contains the core handling logic for the thermal camera application. It includes functions to handle commands received from WebSockets and packets received from C pipes.

# Data from WebSockets

## HandleSetZoomLevel
This function handles the `SetZoomLevel` command, which sets the zoom level of the thermal camera.

```go
func HandleSetZoomLevel(level int32) {
	log.Println("SetZoomLevel command received with level", level)
	SendPacketToC(SET_ZOOM_LEVEL, level)
}
```

## HandleSetColorScheme
This function handles the `SetColorScheme` command, which sets the color scheme of the thermal camera.

```go
func HandleSetColorScheme(scheme thermalcamera.ColorScheme) {
	log.Println("SetColorScheme command received with scheme", thermalcamera.ColorScheme_name[int32(scheme)])
	SendPacketToC(SET_COLOR_SCHEME, int32(scheme))
}
```

# Data from C Pipes

## HandlePacketsFromC
This function handles packets received from C and broadcasts them. It includes handling for various packet types, including setting the zoom level, color scheme, and battery charge level.

```go
func HandlePacketsFromC(cm *ConnectionManager, defaultState *DefaultState) error {
	// ... (code snippet)
}
```

# Channel Data Flow
The data flow in `handlers.go` can be summarized as follows:

1. **WebSockets to Go Server**: Commands like `SetZoomLevel` and `SetColorScheme` are received from the client via WebSockets and handled by the corresponding functions.

2. **Go Server to C Server**: The handled commands are then sent to the C server using named pipes.

3. **C Server to Go Server**: Packets received from the C server are handled and processed. This includes updating the zoom level, color scheme, and battery charge level.

4. **Go Server to Clients**: The processed data is then broadcasted to all connected clients using WebSockets.
