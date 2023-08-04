# Table of Contents
- [Overview](#overview)
- [Constants and Data Structures](#constants-and-data-structures)
- [Command Handlers](#command-handlers)
  - [Zoom Level Handler](#zoom-level-handler)
  - [Color Scheme Handler](#color-scheme-handler)
- [Command Processing](#command-processing)
- [Thread Functions](#thread-functions)
  - [Handling Commands](#handling-commands)
  - [Updating Charge](#updating-charge)
- [Main Function](#main-function)
- [Channel Data Flow](#channel-data-flow)

# Overview
The code snippet represents a multi-threaded C program that communicates with other processes through named pipes. It handles commands to set zoom levels and color schemes for a thermal camera and periodically sends updates on the charge level.

# Constants and Data Structures
```c
#define PIPE_NAME_TO_C "/tmp/toC"
#define PIPE_NAME_FROM_C "/tmp/fromC"
#define SET_ZOOM_LEVEL 1
#define SET_COLOR_SCHEME 2
#define CHARGE_PACKET 3
#define PayloadSize 64

typedef struct {
  uint32_t id;
  char payload[PayloadSize];
} Packet;
```
- Named pipes for communication to and from the C process.
- Command IDs for setting zoom level, color scheme, and sending charge level.
- Packet structure containing an ID and payload.

# Command Handlers

## Zoom Level Handler
```c
void handleZoomLevel(Packet *packet) {
  int32_t zoomLevel = *(int32_t *)packet->payload;
  printf("[C] SetZoomLevel command received with level: %d\n", zoomLevel);
  *(int32_t *)packet->payload = zoomLevel;
}
```
- Extracts the zoom level from the payload.
- Prints the received zoom level.
- Assigns the same value back as a dummy operation.

## Color Scheme Handler
```c
void handleColorScheme(Packet *packet) {
  int32_t colorScheme = *(int32_t *)packet->payload;
  printf("[C] SetColorScheme command received with scheme: %d\n", colorScheme);
  *(int32_t *)packet->payload = colorScheme;
}
```
- Extracts the color scheme from the payload.
- Prints the received color scheme.
- Assigns the same value back as a dummy operation.

# Command Processing
```c
void processCommand(Packet *packet) {
  switch (packet->id) {
  case SET_ZOOM_LEVEL:
    handleZoomLevel(packet);
    break;
  case SET_COLOR_SCHEME:
    handleColorScheme(packet);
    break;
  }
  fflush(stdout);
}
```
- Calls the appropriate handler based on the command ID.
- Flushes the stdout buffer to print log messages immediately.

# Thread Functions

## Handling Commands
```c
void *handleCommands(void *args) {
  // ...
}
```
- Opens named pipes for reading commands and writing responses.
- Reads command packets from the pipe.
- Processes the command using the appropriate handler.
- Writes the response packet back to the pipe.
- Closes the pipes.

## Updating Charge
```c
void *updateCharge(void *args) {
  // ...
}
```
- Opens the named pipe for writing charge updates.
- Generates random charge values and writes them to the pipe.
- Closes the write pipe.

# Main Function
```c
int main() {
  // ...
}
```
- Removes any existing named pipes.
- Creates new named pipes with read/write permissions.
- Creates threads to handle incoming commands and update the charge.
- Waits for the threads to terminate.
- Removes the named pipes.

# Channel Data Flow
1. **Commands to C Process**: Commands are sent to the C process through the named pipe `/tmp/toC`. These include setting zoom levels and color schemes.
2. **Responses from C Process**: Responses are sent from the C process through the named pipe `/tmp/fromC`. These include acknowledgments of the received commands.
3. **Charge Updates**: Periodic charge updates are sent from the C process through the named pipe `/tmp/fromC`.
4. **Synchronization**: A mutex (`pipe_mutex`) is used to synchronize access to the pipes among different threads.
