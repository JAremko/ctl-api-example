package main

import (
	"log"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/JAremko/ctl-api-example/thermalcamera"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow connection from any origin
}

func handleTimeStream(conn *websocket.Conn) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		payload := &thermalcamera.StreamTimeResponse{
			Time: time.Now().Format(time.RFC3339),
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

func echo(w http.ResponseWriter, r *http.Request) {
    log.Println("Incoming connection from:", r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

    log.Println("Upgraded to websocket connection")

	go handleTimeStream(conn)

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

		switch x := payload.MessageType.(type) {
		case *thermalcamera.Command_SetZoomLevel:
			log.Println("SetZoomLevel command received with level", x.SetZoomLevel.Level)
		case *thermalcamera.Command_SetColorScheme:
			log.Println("SetColorScheme command received with scheme", x.SetColorScheme.Scheme)
		case *thermalcamera.Command_GetHumidity:
			log.Println("GetHumidity command received")
		}

		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Println(err)
			return
		}
	}
}

func main() {
	http.HandleFunc("/", echo)
	log.Fatal(http.ListenAndServe(":8085", nil))
}
