# Table of Contents
- [cinterface.go](#cinterfacego)
  - [Package and Imports](#package-and-imports)
  - [Constants and Type Definitions](#constants-and-type-definitions)
  - [Initialization of Pipes](#initialization-of-pipes)
  - [Receiving Packets from C](#receiving-packets-from-c)
  - [Sending Packets to C](#sending-packets-to-c)
  - [Closing Pipes](#closing-pipes)
  - [Channel Data Flow](#channel-data-flow)

# cinterface.go

The `cinterface.go` file is responsible for managing the communication between the Go server and the C server. It uses named pipes for inter-process communication (IPC) and defines the structure and functions to send and receive packets.

## Package and Imports

```go
// Package declaration
package main

import (
	"encoding/binary" // Used for binary data encoding
	"os"              // Used for file handling
	"time"            // Import time for sleep
)
```

The file imports three packages:
- `encoding/binary`: For encoding and decoding binary data.
- `os`: For file handling, specifically for managing named pipes.
- `time`: For controlling sleep time between retries when opening pipes.

## Constants and Type Definitions

```go
const (
	PIPE_NAME_TO_C   = "/tmp/toC"
	PIPE_NAME_FROM_C = "/tmp/fromC"
	SET_ZOOM_LEVEL   = 1
	SET_COLOR_SCHEME = 2
	CHARGE_PACKET    = 3
	PayloadSize      = 64
)

type Packet struct {
	ID      uint32
	Payload [PayloadSize]byte
}
```

- Named pipes are defined with paths `/tmp/toC` and `/tmp/fromC`.
- Constants for different packet types are defined, such as `SET_ZOOM_LEVEL`, `SET_COLOR_SCHEME`, and `CHARGE_PACKET`.
- `Packet` struct represents the structure of a communication packet, with an ID and a payload.

## Initialization of Pipes

```go
func initPipes() {
	// ...
}
```

The `initPipes` function initializes the named pipes for communication with the C program. It opens the pipes for reading and writing, with retries in case of failure.

## Receiving Packets from C

```go
func ReceivePacketFromC() (*Packet, error) {
	// ...
}
```

The `ReceivePacketFromC` function reads a packet from the C program through the named pipe. It extracts the ID and payload from the received data and returns a `Packet` struct.

## Sending Packets to C

```go
func SendPacketToC(packetID uint32, value int32) error {
	// ...
}
```

The `SendPacketToC` function sends a packet to the C program through the named pipe. It constructs the packet with the given ID and value, then writes it to the pipe.

## Closing Pipes

```go
func closePipes() {
	pipeToC.Close()
	pipeFromC.Close()
}
```

The `closePipes` function closes both named pipes, releasing the resources.

## Channel Data Flow

1. **Initialization**: The named pipes are initialized using `initPipes`.
2. **Sending to C**: The Go program sends commands to the C program using `SendPacketToC`. This could include setting zoom levels or color schemes.
3. **Receiving from C**: The Go program receives data from the C program using `ReceivePacketFromC`. This could include receiving charge levels or other information.
4. **Closing**: The pipes are closed using `closePipes` when communication is no longer needed.
