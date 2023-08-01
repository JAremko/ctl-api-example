package main

import (
	"encoding/binary"
	"log"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/JAremko/ctl-api-example/thermalcamera"
)

// WriteRequest is used to wrap WebSocket messages.
type WriteRequest struct {
	messageType int
	data        []byte
}

// ConnectionWrapper encapsulates a WebSocket connection and a channel for write requests.
type ConnectionWrapper struct {
	conn         *websocket.Conn
	writeChannel chan WriteRequest
	stopChannel  chan struct{} // Channel to signal stopping of goroutines
}

// WriteHandler listens for WriteRequests and writes them to the WebSocket.
func (cw *ConnectionWrapper) WriteHandler() {
	for {
		select {
		case <-cw.stopChannel: // If stop signal received, exit goroutine
			return
		case writeReq, ok := <-cw.writeChannel:
			if !ok {
				return // Channel was closed, so exit
			}
			if err := cw.conn.WriteMessage(writeReq.messageType, writeReq.data); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

// handlePacketsFromC reads packets from C and sends them as WebSocket messages.
func handlePacketsFromC(cw *ConnectionWrapper) error {
	for {
		packet, err := ReceivePacketFromC()
		if err != nil {
			log.Println("Error receiving packet:", err)
			return err
		}

		switch packet.ID {
		case CHARGE_PACKET:
			handleChargePacket(cw, packet)
		// You can add more cases here to handle other packet types
		default:
			log.Println("Unknown packet ID:", packet.ID)
		}
	}
}

// handleChargePacket handles a charge packet, constructs a response payload and sends it as a WebSocket message.
func handleChargePacket(cw *ConnectionWrapper, packet *Packet) {
	payload := &thermalcamera.StreamChargeResponse{
		Charge: int32(binary.LittleEndian.Uint32(packet.Payload[:])),
	}
	message, err := proto.Marshal(payload)
	if err != nil {
		log.Println("Error marshaling payload:", err)
		return
	}

	// Safely send to writeChannel or return if stop signal received
	select {
	case <-cw.stopChannel:
		// Connection was closed, so return without writing
		return
	case cw.writeChannel <- WriteRequest{websocket.BinaryMessage, message}:
		// Successfully sent
	}
}

// handleSetZoomLevel handles the SetZoomLevel command.
func handleSetZoomLevel(level int32) {
	log.Println("SetZoomLevel command received with level", level)
	SendPacketToC(SET_ZOOM_LEVEL, level)
}

// handleSetColorScheme handles the SetColorScheme command.
func handleSetColorScheme(scheme thermalcamera.ColorScheme) {
	log.Println("SetColorScheme command received with scheme", thermalcamera.ColorScheme_name[int32(scheme)])
	SendPacketToC(SET_COLOR_SCHEME, int32(scheme))
}

// handleConnection manages a WebSocket connection, including reading and writing messages.
func handleConnection(conn *websocket.Conn) {
	// Initialize ConnectionWrapper with channels
	cw := &ConnectionWrapper{conn: conn, writeChannel: make(chan WriteRequest), stopChannel: make(chan struct{})}
	defer func() {
		conn.WriteMessage(websocket.CloseMessage, []byte{}) // Explicitly send a close message
		conn.Close()
		close(cw.stopChannel)  // Close stopChannel to signal stopping of related goroutines
		close(cw.writeChannel)
	}()

	log.Println("Upgraded to WebSocket connection")

	errorChannel := make(chan error, 1)
	go cw.WriteHandler()
	// Run the packets handler in a new goroutine, listening for errors
	go func() {
		if err := handlePacketsFromC(cw); err != nil {
			errorChannel <- err
		}
	}()

	// Main loop to handle incoming messages from the WebSocket connection
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
			switch messageType {
			case websocket.BinaryMessage:
				var payload thermalcamera.Command
				err = proto.Unmarshal(message, &payload)
				if err != nil {
					log.Println("Error unmarshaling payload:", err)
					continue
				}

				// Switch on the specific command type and handle it
				switch x := payload.CommandType.(type) {
				case *thermalcamera.Command_SetZoomLevel:
					handleSetZoomLevel(int32(x.SetZoomLevel.Level))
				case *thermalcamera.Command_SetColorScheme:
					handleSetColorScheme(x.SetColorScheme.Scheme)
				}
			case websocket.TextMessage:
				log.Println("Received text message:", string(message))
			case websocket.CloseMessage:
				log.Println("Received close message")
				return
			case websocket.PingMessage:
				log.Println("Received ping message")
				conn.WriteMessage(websocket.PongMessage, []byte{})
			case websocket.PongMessage:
				log.Println("Received pong message")
			default:
				log.Println("Unknown message type:", messageType)
			}
		}
	}
}

// echo upgrades HTTP requests to WebSocket connections.
func echo(w http.ResponseWriter, r *http.Request) {
	log.Println("Incoming connection from:", r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	go handleConnection(conn) // Start handling the connection in a new goroutine
}

// upgrader is used to upgrade HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func main() {
	initPipes()              // Initialize the pipes for communication with C
	defer closePipes()       // Ensure that the pipes are closed when the main function returns
	http.HandleFunc("/", echo) // Register the echo handler for HTTP requests
	log.Fatal(http.ListenAndServe(":8085", nil)) // Start the HTTP server on port 8085
}
