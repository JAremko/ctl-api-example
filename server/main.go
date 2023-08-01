package main

import (
	"encoding/binary"
	"log"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/JAremko/ctl-api-example/thermalcamera"
)

// WriteRequest wraps the WebSocket message type and data.
type WriteRequest struct {
	messageType int
	data        []byte
}

// ConnectionWrapper wraps a WebSocket connection and a channel for write requests.
type ConnectionWrapper struct {
	conn         *websocket.Conn
	writeChannel chan WriteRequest
}

// WriteHandler processes write requests for the WebSocket connection.
func (cw *ConnectionWrapper) WriteHandler() {
	for writeReq := range cw.writeChannel {
		if err := cw.conn.WriteMessage(writeReq.messageType, writeReq.data); err != nil {
			log.Println(err)
			return
		}
	}
}

// handleChargeStream reads packets from C and sends them as WebSocket messages.
func handleChargeStream(cw *ConnectionWrapper) {
	for {
		packet, err := ReceivePacketFromC()
		if err != nil {
			log.Println("Go Error receiving packet:", err)
			return
		}

		payload := &thermalcamera.StreamChargeResponse{
			Charge: int32(binary.LittleEndian.Uint32(packet.Payload[:])),
		}
		message, err := proto.Marshal(payload)
		if err != nil {
			log.Println("Go Error marshaling payload:", err)
			return
		}
		cw.writeChannel <- WriteRequest{websocket.BinaryMessage, message}
	}
}

// handleConnection manages a WebSocket connection, including reading and writing messages.
func handleConnection(conn *websocket.Conn) {
	cw := &ConnectionWrapper{conn: conn, writeChannel: make(chan WriteRequest)}
	defer conn.Close()
	defer close(cw.writeChannel)

	log.Println("Go: Upgraded to websocket connection")

	go cw.WriteHandler()          // Start the write handler in a new goroutine.
	go handleChargeStream(cw) // Start the charge stream handler in a new goroutine.

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var payload thermalcamera.Command
		err = proto.Unmarshal(message, &payload)
		if err != nil {
			log.Println("Go Error unmarshaling payload:", err)
			continue
		}

		switch x := payload.CommandType.(type) { // Handle specific command types.
		case *thermalcamera.Command_SetZoomLevel:
			log.Println("Go SetZoomLevel command received with level", x.SetZoomLevel.Level)
			SendPacketToC(SET_ZOOM_LEVEL, int32(x.SetZoomLevel.Level))
		case *thermalcamera.Command_SetColorScheme:
			log.Println("Go SetColorScheme command received with scheme",
				thermalcamera.ColorScheme_name[int32(x.SetColorScheme.Scheme)])
			SendPacketToC(SET_COLOR_SCHEME, int32(x.SetColorScheme.Scheme))
		}

		cw.writeChannel <- WriteRequest{messageType, message}
	}
}

// echo handles HTTP requests by upgrading them to WebSocket connections.
func echo(w http.ResponseWriter, r *http.Request) {
	log.Println("Go: Incoming connection from:", r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	go handleConnection(conn) // Start handling the connection in a new goroutine.
}

// upgrader is used to upgrade HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow connections from any origin.
}

func main() {
	initPipes()              // Initialize the named pipes for communication with C.
	defer closePipes()       // Ensure that the pipes are closed when the main function returns.

	http.HandleFunc("/", echo) // Register the echo handler for HTTP requests.
	log.Fatal(http.ListenAndServe(":8085", nil)) // Start the HTTP server on port 8085.
}
