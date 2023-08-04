### Table of Contents
- [Introduction](#introduction)
- [Imports](#imports)
- [Struct Definitions](#struct-definitions)
- [Functions](#functions)
  - [sendDefaultState](#senddefaultstate)
  - [Broadcast](#broadcast)
  - [AddConnection](#addconnection)
  - [RemoveConnection](#removeconnection)
  - [WriteHandler](#writehandler)
  - [handleConnection](#handleconnection)
  - [echo](#echo)
- [Main Function](#main-function)
- [Data Flow](#data-flow)

### Introduction
This file contains the main server logic for handling WebSocket connections and communication with a thermal camera. It includes functions for managing connections, broadcasting messages, handling incoming messages, and more.

### Imports
```go
import (
	"log"
	"net/http"
	"sync"

	"github.com/JAremko/ctl-api-example/thermalcamera"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)
```
- `log`: For logging messages.
- `net/http`: To implement the HTTP server.
- `sync`: To handle synchronization.
- `thermalcamera`: Importing Protobuf definitions.
- `proto`: Protobuf encoding and decoding package.
- `websocket`: Package for WebSocket implementation.

### Struct Definitions
```go
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
	mutex       sync.Mutex
}
```
- `WriteRequest`: Encapsulates a write request with message type and data.
- `ConnectionWrapper`: Wraps a WebSocket connection and channels for writing and stopping.
- `ConnectionManager`: Keeps track of active WebSocket connections.

### Functions

#### sendDefaultState
Sends the default state of the thermal camera to the client.
```go
func sendDefaultState(cw *ConnectionWrapper, defaultState *DefaultState) {
	// ...
}
```

#### Broadcast
Sends a message to all active connections.
```go
func (cm *ConnectionManager) Broadcast(writeReq WriteRequest) {
	// ...
}
```

#### AddConnection
Adds a connection to the manager.
```go
func (cm *ConnectionManager) AddConnection(connection *ConnectionWrapper) {
	// ...
}
```

#### RemoveConnection
Removes a connection from the manager.
```go
func (cm *ConnectionManager) RemoveConnection(connection *ConnectionWrapper) {
	// ...
}
```

#### WriteHandler
Deals with writing messages to the WebSocket.
```go
func (cw *ConnectionWrapper) WriteHandler(errorChannel chan error) {
	// ...
}
```

#### handleConnection
Manages the life cycle of a WebSocket connection.
```go
func handleConnection(conn *websocket.Conn, cm *ConnectionManager, defaultState *DefaultState) {
	// ...
}
```

#### echo
Handles incoming HTTP connections and upgrades them to WebSockets.
```go
func echo(w http.ResponseWriter, r *http.Request, cm *ConnectionManager, defaultState *DefaultState) {
	// ...
}
```

### Main Function
The main function initializes the program, sets up the default state, initializes the connection manager, and starts the HTTP server.
```go
func main() {
	// ...
}
```

### Data Flow
1. **HTTP Connection**: Clients connect via HTTP, and the connection is upgraded to WebSocket using the `echo` function.
2. **Connection Handling**: Each connection is handled in a separate goroutine by `handleConnection`, which manages the life cycle of the connection.
3. **Writing to WebSocket**: Writing to the WebSocket is handled by `WriteHandler`, which listens for write requests on the `writeChannel`.
4. **Broadcasting Messages**: Messages can be broadcasted to all active connections using the `Broadcast` method of the `ConnectionManager`.
5. **Handling Commands**: Incoming commands are read and unmarshaled, and appropriate handlers are called based on the payload.
6. **Stopping Connections**: Connections can be stopped by sending a signal on the `stopChannel`.
