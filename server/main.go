package main

import (
	"encoding/binary"
	"log"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/JAremko/ctl-api-example/thermalcamera"
)

type WriteRequest struct {
	messageType int
	data        []byte
}

type ConnectionWrapper struct {
	conn         *websocket.Conn
	writeChannel chan WriteRequest
}

func (cw *ConnectionWrapper) WriteHandler() {
	for writeReq := range cw.writeChannel {
		if err := cw.conn.WriteMessage(writeReq.messageType, writeReq.data); err != nil {
			log.Println(err)
			return
		}
	}
}

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

func handleConnection(conn *websocket.Conn) {
	cw := &ConnectionWrapper{conn: conn, writeChannel: make(chan WriteRequest)}
	defer conn.Close()
	defer close(cw.writeChannel)

	log.Println("Go: Upgraded to websocket connection")

	go cw.WriteHandler()
	go handleChargeStream(cw)

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

		switch x := payload.CommandType.(type) {
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

func echo(w http.ResponseWriter, r *http.Request) {
	log.Println("Go: Incoming connection from:", r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	go handleConnection(conn)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func main() {
	initPipes()
	defer closePipes()

	http.HandleFunc("/", echo)
	log.Fatal(http.ListenAndServe(":8085", nil))
}
