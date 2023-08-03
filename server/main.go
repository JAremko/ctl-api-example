package main

import (
	"encoding/binary" // Package for handling binary data
	"log"             // Logging package
	"net/http"        // Package for HTTP server implementation
	"sync"            // Package to handle synchronization

	"github.com/JAremko/ctl-api-example/thermalcamera" // Importing Protobuf definitions
	"github.com/golang/protobuf/proto"                 // Protobuf encoding and decoding package
	"github.com/gorilla/websocket"                     // Package for WebSocket implementation
)

// WriteRequest is a struct for encapsulating write request
type WriteRequest struct {
	messageType int
	data        []byte
}

// ConnectionWrapper wraps WebSocket connection and channels for writing and stopping
type ConnectionWrapper struct {
	conn         *websocket.Conn
	writeChannel chan WriteRequest
	stopChannel  chan struct{}
}

// ConnectionManager keeps track of active WebSocket connections
type ConnectionManager struct {
	connections map[*ConnectionWrapper]bool
}

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

// sendDefaultState sends the default state of the thermal camera to the client
func sendDefaultState(cw *ConnectionWrapper, defaultState *DefaultState) {
	// Creating and marshaling a payload message
	payload := &thermalcamera.Payload{
		SetZoomLevel:   &thermalcamera.SetZoomLevel{Level: defaultState.GetZoomLevel()},
		SetColorScheme: &thermalcamera.SetColorScheme{Scheme: defaultState.GetColorScheme()},
		AccChargeLevel: &thermalcamera.AccChargeLevel{Charge: defaultState.GetBatteryLevel()},
	}
	message, err := proto.Marshal(payload)
	if err != nil {
		log.Println("Error marshaling default state:", err)
		return
	}
	// Sending the message through the write channel
	cw.writeChannel <- WriteRequest{websocket.BinaryMessage, message}
}

// Broadcast sends a message to all active connections
func (cm *ConnectionManager) Broadcast(writeReq WriteRequest) {
	for connection := range cm.connections {
		select {
		case connection.writeChannel <- writeReq:
		default:
			close(connection.stopChannel)
			close(connection.writeChannel)
			delete(cm.connections, connection)
		}
	}
}

// AddConnection adds a connection to the manager
func (cm *ConnectionManager) AddConnection(connection *ConnectionWrapper) {
	cm.connections[connection] = true
}

// RemoveConnection removes a connection from the manager
func (cm *ConnectionManager) RemoveConnection(connection *ConnectionWrapper) {
	delete(cm.connections, connection)
}

// WriteHandler deals with writing messages to the WebSocket
func (cw *ConnectionWrapper) WriteHandler() {
	for {
		select {
		case <-cw.stopChannel: // Return if stop signal received
			return
		case writeReq, ok := <-cw.writeChannel: // Handle write requests
			if !ok {
				return
			}
			if err := cw.conn.WriteMessage(writeReq.messageType, writeReq.data); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

// handleSetZoomLevel handles the SetZoomLevel command
func handleSetZoomLevel(level int32) {
	log.Println("SetZoomLevel command received with level", level)
	SendPacketToC(SET_ZOOM_LEVEL, level)
}

// handleSetColorScheme handles the SetColorScheme command
func handleSetColorScheme(scheme thermalcamera.ColorScheme) {
	log.Println("SetColorScheme command received with scheme", thermalcamera.ColorScheme_name[int32(scheme)])
	SendPacketToC(SET_COLOR_SCHEME, int32(scheme))
}

// handlePacketsFromC handles packets received from C and broadcasts them
func handlePacketsFromC(cm *ConnectionManager, defaultState *DefaultState) error {
	for {
		// Handling various packets received from C
		packet, err := ReceivePacketFromC()
		if err != nil {
			log.Println("Error receiving packet:", err)
			return err
		}
		var payload thermalcamera.Payload
		switch packet.ID {
		case SET_ZOOM_LEVEL:
			zoomLevel := int32(binary.LittleEndian.Uint32(packet.Payload[:4]))
			payload.SetZoomLevel = &thermalcamera.SetZoomLevel{
				Level: zoomLevel,
			}
			defaultState.UpdateZoomLevel(zoomLevel) // Update zoom level
		case SET_COLOR_SCHEME:
			colorScheme := thermalcamera.ColorScheme(binary.LittleEndian.Uint32(packet.Payload[:4]))
			payload.SetColorScheme = &thermalcamera.SetColorScheme{
				Scheme: colorScheme,
			}
			defaultState.UpdateColorScheme(colorScheme) // Update color scheme
		case CHARGE_PACKET:
			charge := int32(binary.LittleEndian.Uint32(packet.Payload[:]))
			payload.AccChargeLevel = &thermalcamera.AccChargeLevel{
				Charge: charge,
			}
			defaultState.UpdateBatteryLevel(charge) // Update battery level
		default:
			log.Println("Unknown packet ID:", packet.ID)
		}
		message, err := proto.Marshal(&payload)
		if err != nil {
			log.Println("Error marshaling payload:", err)
			return err
		}
		cm.Broadcast(WriteRequest{websocket.BinaryMessage, message})
	}
}

// handleConnection manages the life cycle of a WebSocket connection
func handleConnection(conn *websocket.Conn, cm *ConnectionManager, defaultState *DefaultState) {
	cw := &ConnectionWrapper{conn: conn, writeChannel: make(chan WriteRequest), stopChannel: make(chan struct{})}
	cm.AddConnection(cw)
	defer func() {
		conn.WriteMessage(websocket.CloseMessage, []byte{})
		conn.Close()
		close(cw.stopChannel)
		close(cw.writeChannel)
		cm.RemoveConnection(cw)
	}()
	log.Println("Upgraded to WebSocket connection")

	errorChannel := make(chan error, 1)
	go cw.WriteHandler()

	sendDefaultState(cw, defaultState)

	for {
		select {
		case err := <-errorChannel:
			log.Println("Error in stream handling:", err)
			return

		default:
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			if messageType == websocket.BinaryMessage {
				var payload thermalcamera.Payload
				err = proto.Unmarshal(message, &payload)
				if err != nil {
					log.Println("Error unmarshaling payload:", err)
					continue
				}

				// Handling based on fields present in the payload
				if payload.SetZoomLevel != nil {
					handleSetZoomLevel(int32(payload.SetZoomLevel.Level))
				}
				if payload.SetColorScheme != nil {
					handleSetColorScheme(payload.SetColorScheme.Scheme)
				}
			}
		}
	}
}

// echo handles incoming HTTP connections and upgrades them to WebSockets
func echo(w http.ResponseWriter, r *http.Request, cm *ConnectionManager, defaultState *DefaultState) {
	log.Println("Incoming connection from:", r.RemoteAddr)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	go handleConnection(conn, cm, defaultState) // Starting connection handling in a new goroutine
}

// upgrader is used to upgrade HTTP connections to WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// main function initializes the program
func main() {
	initPipes()
	defer closePipes()

	// Setting default state
	defaultState := &DefaultState{
		ZoomLevel:    1,
		ColorScheme:  thermalcamera.ColorScheme_BLACK_HOT,
		BatteryLevel: 100,
	}

	// Initializing connection manager
	connectionManager := &ConnectionManager{
		connections: make(map[*ConnectionWrapper]bool),
	}

	// Starting a goroutine to handle packets from C
	go func() {
		if err := handlePacketsFromC(connectionManager, defaultState); err != nil {
			log.Println("Error in stream handling:", err)
		}
	}()

	// Setting up HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		echo(w, r, connectionManager, defaultState)
	})
	log.Fatal(http.ListenAndServe(":8085", nil)) // Starting the server on port 8085
}
