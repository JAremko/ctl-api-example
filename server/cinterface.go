// Package declaration
package main

import (
	"encoding/binary" // Used for binary data encoding
	"log"             // Used for logging
	"os"              // Used for file handling
)

// Constants used for defining various communication properties
const (
	PIPE_NAME_TO_C   = "/tmp/toC"   // Name of the pipe to send data to the C program
	PIPE_NAME_FROM_C = "/tmp/fromC" // Name of the pipe to receive data from the C program
	SET_ZOOM_LEVEL   = 1            // ID for the command to set the zoom level
	SET_COLOR_SCHEME = 2            // ID for the command to set the color scheme
	CHARGE_PACKET    = 3            // ID for the charge packet
	PayloadSize      = 64           // Size of the payload in bytes
)

// Packet represents the structure of a communication packet
type Packet struct {
	ID      uint32            // ID to uniquely identify the type of packet
	Payload [PayloadSize]byte // Payload to carry the data of the packet
}

var pipeToC *os.File   // File handle for the named pipe to send data to the C program
var pipeFromC *os.File // File handle for the named pipe to receive data from the C program

// initPipes initializes the named pipes for communication with the C program.
func initPipes() {
	var err error
	pipeFromC, err = os.Open(PIPE_NAME_FROM_C) // Open the named pipe for reading from the C program
	if err != nil {
		log.Fatal(err) // Log the error and terminate if opening fails
	}
	pipeToC, err = os.OpenFile(PIPE_NAME_TO_C, os.O_WRONLY, os.ModeNamedPipe) // Open the named pipe for writing to the C program
	if err != nil {
		log.Fatal(err) // Log the error and terminate if opening fails
	}
}

// ReceivePacketFromC reads a packet from the C program through the named pipe.
func ReceivePacketFromC() (*Packet, error) {
	var packet Packet
	// Buffer to hold the data; 4 bytes for uint32 ID to uniquely identify the type of packet, and PayloadSize for the actual payload data
	var buf [4 + PayloadSize]byte
	if _, err := pipeFromC.Read(buf[:]); err != nil { // Read from the pipe
		pipeFromC.Close()
		pipeFromC, _ = os.Open(PIPE_NAME_FROM_C) // Reopen the pipe if an error occurs
		return nil, err
	}
	packet.ID = binary.LittleEndian.Uint32(buf[:4]) // Extract the 4-byte ID
	copy(packet.Payload[:], buf[4:])                // Copy the rest as payload
	return &packet, nil
}

// SendPacketToC sends a packet to the C program through the named pipe.
func SendPacketToC(packetID uint32, value int32) error {
	var packet Packet
	packet.ID = packetID                                             // Set the packet ID
	binary.LittleEndian.PutUint32(packet.Payload[:4], uint32(value)) // Store the 4-byte value in the payload
	// Buffer to hold the data, including 4 bytes for the uint32 ID, and PayloadSize for the payload
	var buf [4 + PayloadSize]byte
	binary.LittleEndian.PutUint32(buf[:4], packet.ID) // Place the ID in the buffer
	copy(buf[4:], packet.Payload[:])                  // Copy the payload into the buffer
	_, err := pipeToC.Write(buf[:])                   // Write to the named pipe
	if err != nil {
		pipeToC.Close()
		pipeToC, err = os.OpenFile(PIPE_NAME_TO_C, os.O_WRONLY, os.ModeNamedPipe) // Reopen the pipe if an error occurs
		if err != nil {
			return err // Return error if reopening fails
		}
	}
	return nil // Return no error if successful
}

// closePipes closes both named pipes, releasing the resources.
func closePipes() {
	pipeToC.Close()   // Close the named pipe to C
	pipeFromC.Close() // Close the named pipe from C
}
