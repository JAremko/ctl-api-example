package main

import (
	"encoding/binary"
	"log"
	"net/http"
	"os"
	"unsafe"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/JAremko/ctl-api-example/thermalcamera"
)

const (
	PIPE_NAME_TO_C      = "/tmp/toC"
	PIPE_NAME_FROM_C    = "/tmp/fromC"
	SET_ZOOM_LEVEL      = 1
	SET_COLOR_SCHEME    = 2
	STREAM_CHARGE_RESPONSE = 3
)

type Packet struct {
	ID      uint32
	Payload [4]byte
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func handleChargeStream(conn *websocket.Conn) {
	for {
		var packet Packet
		err := receivePacketFromC(&packet)
		if err != nil {
			log.Println("Error receiving packet:", err)
			return
		}

		payload := &thermalcamera.StreamChargeResponse{
			Charge: int32(binary.LittleEndian.Uint32(packet.Payload[:])),
		}
		message, err := proto.Marshal(payload)
		if err != nil {
			log.Println("Error marshaling payload:", err)
			return
		}
		if err := conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
			log.Println(err)
			return
		}
	}
}

func receivePacketFromC(packet *Packet) error {
	pipeFromC, err := os.Open(PIPE_NAME_FROM_C)
	if err != nil {
		return err
	}
	defer pipeFromC.Close()

	var buf [8]byte
	if _, err := pipeFromC.Read(buf[:]); err != nil {
		return err
	}
	packet.ID = binary.LittleEndian.Uint32(buf[:4])
	copy(packet.Payload[:], buf[4:])

	return nil
}

func handleConnection(conn *websocket.Conn) {
	defer conn.Close()

	log.Println("Upgraded to websocket connection")

	pipeToC, err := os.OpenFile(PIPE_NAME_TO_C, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		log.Println(err)
		return
	}
	defer pipeToC.Close()

	go handleChargeStream(conn)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var payload thermalcamera.Command
		err = proto.Unmarshal(message, &payload)
		if err != nil {
			log.Println("Error unmarshaling payload:", err)
			continue
		}

		switch x := payload.CommandType.(type) {
		case *thermalcamera.Command_SetZoomLevel:
			log.Println("SetZoomLevel command received with level", x.SetZoomLevel.Level)
			sendPacketToC(pipeToC, SET_ZOOM_LEVEL, int32(x.SetZoomLevel.Level))
		case *thermalcamera.Command_SetColorScheme:
			log.Println("SetColorScheme command received with scheme",
				thermalcamera.ColorScheme_name[int32(x.SetColorScheme.Scheme)])
			sendPacketToC(pipeToC, SET_COLOR_SCHEME, int32(x.SetColorScheme.Scheme))
		}

		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Println(err)
			return
		}
	}
}

func sendPacketToC(pipeToC *os.File, packetID uint32, value int32) {
	var packet Packet
	packet.ID = packetID
	binary.LittleEndian.PutUint32(packet.Payload[:], uint32(value))

	if _, err := pipeToC.Write((*[8]byte)(unsafe.Pointer(&packet))[:]); err != nil {
		log.Println(err)
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	log.Println("Incoming connection from:", r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	go handleConnection(conn)
}

func main() {
	http.HandleFunc("/", echo)
	log.Fatal(http.ListenAndServe(":8085", nil))
}
