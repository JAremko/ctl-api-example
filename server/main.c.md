# Readme.md

## Table of Contents

- [Overview](#overview)
- [Data Flow](#data-flow)
- [Functions](#functions)
    - [handleZoomLevel](#handlezoomlevel)
    - [handleColorScheme](#handlecolorscheme)
    - [processCommand](#processcommand)
    - [readUntilDelimiter](#readuntildelimiter)
    - [drainPipe](#drainpipe)
    - [handleCommands](#handlecommands)
    - [updateCharge](#updatecharge)
    - [main](#main)

## Overview

This C program acts as an intermediary that communicates with a Go process via named pipes. The program processes commands from the Go program and updates a simulated charge status back to it. The code uses COBS (Consistent Overhead Byte Stuffing) for encoding and decoding the data before sending and after receiving it via pipes.

## Data Flow

1. **Initialization**: The main program starts and initializes two named pipes: `/tmp/toC` and `/tmp/fromC`.
2. **Thread Creation**: Two separate threads are created:
    - `commandThread`: Listens for commands from the Go process and responds to them.
    - `chargeThread`: Periodically updates the charge status back to the Go process.
3. **Receiving Commands**: The `commandThread` reads data from the `/tmp/toC` pipe until it encounters a delimiter (0-byte). After decoding the COBS-encoded data, it processes the command.
4. **Sending Responses**: After processing, the data is COBS-encoded again and sent back via the `/tmp/fromC` pipe.
5. **Updating Charge**: The `chargeThread` simulates charge updates by randomly generating a value between 0 and 100, COBS-encoding this data, and sending it back to the Go process.

## Functions

### handleZoomLevel

- Processes the `SET_ZOOM_LEVEL` command.
- Parameters: A `Packet` pointer containing the command.
- Returns: The size of the processed payload.

### handleColorScheme

- Processes the `SET_COLOR_SCHEME` command.
- Parameters: A `Packet` pointer containing the command.
- Returns: The size of the processed payload.

### processCommand

- Processes the command by checking its ID and delegating to the appropriate handler.
- Parameters: A `Packet` pointer containing the command.
- Returns: The size of the processed payload.

### readUntilDelimiter

- Reads data from the given file descriptor until it encounters a delimiter.
- Parameters: File descriptor, buffer to store data, maximum size.
- Returns: Number of bytes read.

### drainPipe

- Drains excess bytes from the pipe if the `MAX_BUFFER_SIZE` is reached without encountering a delimiter.
- Parameters: File descriptor.
- Returns: Number of bytes drained.

### handleCommands

- Main function for the `commandThread`.
- Listens for commands from the Go process and responds to them.

### updateCharge

- Main function for the `chargeThread`.
- Periodically sends a simulated charge status back to the Go process.

### main

- Main entry point for the program.
- Initializes named pipes and threads, then waits for threads to complete.
