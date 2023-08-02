package main

import (
	"encoding/binary"
	"log"
	"net/http"
	"sync"

	"github.com/JAremko/ctl-api-example/thermalcamera"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
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

type DefaultState struct {
    sync.Mutex
    ZoomLevel    int32
    ColorScheme  thermalcamera.ColorScheme
    BatteryLevel int32
}

func (ds *DefaultState) GetZoomLevel() int32 {
    ds.Lock()
    defer ds.Unlock()
    return ds.ZoomLevel
}

func (ds *DefaultState) GetColorScheme() thermalcamera.ColorScheme {
    ds.Lock()
    defer ds.Unlock()
    return ds.ColorScheme
}

func (ds *DefaultState) GetBatteryLevel() int32 {
    ds.Lock()
    defer ds.Unlock()
    return ds.BatteryLevel
}

func (ds *DefaultState) UpdateZoomLevel(level int32) {
    ds.Lock()
    defer ds.Unlock()
    ds.ZoomLevel = level
}

func (ds *DefaultState) UpdateColorScheme(scheme thermalcamera.ColorScheme) {
    ds.Lock()
    defer ds.Unlock()
    ds.ColorScheme = scheme
}

func (ds *DefaultState) UpdateBatteryLevel(level int32) {
    ds.Lock()
    defer ds.Unlock()
    ds.BatteryLevel = level
}

func sendDefaultState(cw *ConnectionWrapper, defaultState *DefaultState) {
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
    cw.writeChannel <- WriteRequest{websocket.BinaryMessage, message}
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

func handleSetZoomLevel(level int32) {
	log.Println("SetZoomLevel command received with level", level)
	SendPacketToC(SET_ZOOM_LEVEL, level)
}

func handleSetColorScheme(scheme thermalcamera.ColorScheme) {
	log.Println("SetColorScheme command received with scheme", thermalcamera.ColorScheme_name[int32(scheme)])
	SendPacketToC(SET_COLOR_SCHEME, int32(scheme))
}

func handlePacketsFromC(cm *ConnectionManager) error {
	for {
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
		case SET_COLOR_SCHEME:
			colorScheme := thermalcamera.ColorScheme(binary.LittleEndian.Uint32(packet.Payload[:4]))
			payload.SetColorScheme = &thermalcamera.SetColorScheme{
				Scheme: colorScheme,
			}
		case CHARGE_PACKET:
			charge := int32(binary.LittleEndian.Uint32(packet.Payload[:]))
			payload.AccChargeLevel = &thermalcamera.AccChargeLevel{
				Charge: charge,
			}
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

func echo(w http.ResponseWriter, r *http.Request, cm *ConnectionManager, defaultState *DefaultState) {
	log.Println("Incoming connection from:", r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	go handleConnection(conn, cm, defaultState)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func main() {
	initPipes()
	defer closePipes()

	defaultState := &DefaultState{
		ZoomLevel:    1,
		ColorScheme:  thermalcamera.ColorScheme_BLACK_HOT,
		BatteryLevel: 100,
	}

	connectionManager := &ConnectionManager{
		connections: make(map[*ConnectionWrapper]bool),
	}

	go func() {
		if err := handlePacketsFromC(connectionManager); err != nil {
			log.Println("Error in stream handling:", err)
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		echo(w, r, connectionManager, defaultState)
	})
	log.Fatal(http.ListenAndServe(":8085", nil))
}
