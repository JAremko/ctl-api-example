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
	stopChannel  chan struct{}
}

type ConnectionManager struct {
	connections map[*ConnectionWrapper]bool
}

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

func (cm *ConnectionManager) AddConnection(connection *ConnectionWrapper) {
	cm.connections[connection] = true
}

func (cm *ConnectionManager) RemoveConnection(connection *ConnectionWrapper) {
	delete(cm.connections, connection)
}

func (cw *ConnectionWrapper) WriteHandler() {
	for {
		select {
		case <-cw.stopChannel:
			return
		case writeReq, ok := <-cw.writeChannel:
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

func handlePacketsFromC(cm *ConnectionManager) error {
	for {
		packet, err := ReceivePacketFromC()
		if err != nil {
			log.Println("Error receiving packet:", err)
			return err
		}

		switch packet.ID {
		case CHARGE_PACKET:
			handleChargePacket(cm, packet)
		default:
			log.Println("Unknown packet ID:", packet.ID)
		}
	}
}

func handleChargePacket(cm *ConnectionManager, packet *Packet) {
	payload := &thermalcamera.StreamChargeResponse{
		Charge: int32(binary.LittleEndian.Uint32(packet.Payload[:])),
	}
	message, err := proto.Marshal(payload)
	if err != nil {
		log.Println("Error marshaling payload:", err)
		return
	}

	cm.Broadcast(WriteRequest{websocket.BinaryMessage, message})
}


func handleSetZoomLevel(level int32) {
	log.Println("SetZoomLevel command received with level", level)
	SendPacketToC(SET_ZOOM_LEVEL, level)
}


func handleSetColorScheme(scheme thermalcamera.ColorScheme) {
	log.Println("SetColorScheme command received with scheme", thermalcamera.ColorScheme_name[int32(scheme)])
	SendPacketToC(SET_COLOR_SCHEME, int32(scheme))
}

func handleConnection(conn *websocket.Conn, cm *ConnectionManager) {
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
	go func() {
		if err := handlePacketsFromC(cm); err != nil {
			errorChannel <- err
		}
	}()

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

func echo(w http.ResponseWriter, r *http.Request, cm *ConnectionManager) {
	log.Println("Incoming connection from:", r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	go handleConnection(conn, cm)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func main() {
	initPipes()
	defer closePipes()

	connectionManager := &ConnectionManager{
		connections: make(map[*ConnectionWrapper]bool),
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		echo(w, r, connectionManager)
	})
	log.Fatal(http.ListenAndServe(":8085", nil))
}
