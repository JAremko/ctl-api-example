package main

import (
	"encoding/binary"
	"log"
	"os"
	"unsafe"
)

const (
	PIPE_NAME_TO_C         = "/tmp/toC"
	PIPE_NAME_FROM_C       = "/tmp/fromC"
	SET_ZOOM_LEVEL         = 1
	SET_COLOR_SCHEME       = 2
	STREAM_CHARGE_RESPONSE = 3
	PayloadSize            = 64
)

type Packet struct {
	ID      uint32
	Payload [PayloadSize]byte
}

var pipeToC *os.File
var pipeFromC *os.File

func initPipes() {
	var err error
	pipeFromC, err = os.Open(PIPE_NAME_FROM_C)
	if err != nil {
		log.Fatal(err)
	}
	pipeToC, err = os.OpenFile(PIPE_NAME_TO_C, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		log.Fatal(err)
	}
}

func ReceivePacketFromC() (*Packet, error) {
	var packet Packet
	var buf [PayloadSize + 4]byte
	if _, err := pipeFromC.Read(buf[:]); err != nil {
		pipeFromC.Close()
		pipeFromC, _ = os.Open(PIPE_NAME_FROM_C)
		return nil, err
	}
	packet.ID = binary.LittleEndian.Uint32(buf[:4])
	copy(packet.Payload[:], buf[4:])
	return &packet, nil
}

func SendPacketToC(packetID uint32, value int32) error {
	var packet Packet
	packet.ID = packetID
	binary.LittleEndian.PutUint32(packet.Payload[:], uint32(value))

	_, err := pipeToC.Write((*[PayloadSize + 4]byte)(unsafe.Pointer(&packet))[:])
	if err != nil {
		pipeToC.Close()
		pipeToC, _ = os.OpenFile(PIPE_NAME_TO_C, os.O_WRONLY, os.ModeNamedPipe)
		return err
	}
	return nil
}

func closePipes() {
	pipeToC.Close()
	pipeFromC.Close()
}
