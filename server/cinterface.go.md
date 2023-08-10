# `cinterface.go` Documentation

`cinterface.go` serves as an interface between Go and C programs. It facilitates the communication between the two via named pipes. The communication is established through packets that have a unique ID and payload. This document provides a comprehensive guide on the structure, functionality, and data flow of `cinterface.go`.

## Table of Contents
- [Overview](#overview)
- [Constants and Types](#constants-and-types)
  - [Constants](#constants)
  - [Packet Structure](#packet-structure)
- [Global Variables](#global-variables)
- [Functions](#functions)
  - [initPipes](#initpipes)
  - [ReceivePacketFromC](#receivepacketfromc)
  - [SendPacketToC](#sendpackettoc)
  - [closePipes](#closepipes)
- [Data Flow](#data-flow)

## Overview
The `cinterface.go` code provides a mechanism to:
1. Initialize communication pipes.
2. Receive data packets from a C program.
3. Send data packets to a C program.
4. Close the communication pipes.

## Constants and Types

### Constants

- `PIPE_NAME_TO_C`: Named pipe for sending data to the C program.
- `PIPE_NAME_FROM_C`: Named pipe for receiving data from the C program.
- `SET_ZOOM_LEVEL`: ID representing a command to set the zoom level.
- `SET_COLOR_SCHEME`: ID representing a command to set the color scheme.
- `CHARGE_PACKET`: ID representing a packet for battery charge info.
- `MaxPayloadSize`: Maximum payload size for a packet.

### Packet Structure

The `Packet` structure defines the blueprint for the packets exchanged between the Go program and the C program.

```go
type Packet struct {
	ID      uint32
	Payload [MaxPayloadSize]byte
}
```

## Global Variables

- `pipeToC`: File pointer for the pipe that sends data to the C program.
- `pipeFromC`: File pointer for the pipe that receives data from the C program.
- `readBuffer`: Buffer that holds the data read from the C program.

## Functions

### initPipes

This function initializes the named pipes required for communication. It waits until both the send and receive pipes are available.

### ReceivePacketFromC

This function reads data from the C program through the named pipe, `PIPE_NAME_FROM_C`. It decodes the received data and constructs a `Packet` from it. If there's an error or the read buffer does not contain a full packet, the function handles it gracefully and continues reading.

### SendPacketToC

This function allows the Go program to send data to the C program. It accepts a packet ID and a value, encodes it into COBS (Consistent Overhead Byte Stuffing) format, and then sends it through the named pipe, `PIPE_NAME_TO_C`.

### closePipes

This function closes the named pipes, effectively ending the communication between the Go and C programs.

## Data Flow

1. The Go program initializes the named pipes using `initPipes`.
2. The Go program waits for data from the C program using `ReceivePacketFromC`. The data is buffered until a complete packet (up to the delimiter) is received.
3. When a complete packet is detected in the buffer, it's decoded from its COBS format.
4. The Go program can send data to the C program using `SendPacketToC`. The data is encoded into COBS format and then sent.
5. To end the communication, the Go program closes the pipes using `closePipes`.
