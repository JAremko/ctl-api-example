# Thermal Camera API Example

## Introduction

This repository contains a client-server application demonstrating the use of Protobufs and WebSockets to control a thermal camera. The server part consists of two components: a Go server and a C server, communicating via named pipes. The client is a JavaScript-based frontend that now includes synchronization through iframes.

## Repository Structure

- `/server`: Contains the Go server code (`main.go` and `cinterface.go`) and C code (`main.c`).
- `/client`: Contains the frontend JavaScript code (`client.js`) and HTML (`ctl.html` and `index.html`).
- `thermalcamera.proto`: The Protobuf schema defining the service and messages.
- `Dockerfile.client` & `Dockerfile.server`: Dockerfiles used to build the client and server containers.
- `docker-compose.yml`: Docker Compose file to orchestrate the client and server containers.

## Protobuf Definitions (`thermalcamera.proto`)

The `thermalcamera.proto` file defines the messages and services for controlling a thermal camera. The definitions include:

- **Payload**: A wrapper message that includes one of the following payload types:
  - **SetZoomLevel**: To set the zoom level of the camera.
  - **SetColorScheme**: To set the color scheme of the camera.
  - **AccChargeLevel**: To get the accumulated charge level of the camera, represented as a percentage.
- **ColorScheme**: An enumeration of available color schemes, including UNKNOWN, SEPIA, BLACK_HOT, and WHITE_HOT.

These definitions are used to serialize the data sent between the client and server, ensuring a consistent and robust communication protocol.

## Usage

### Building and Running the Application

1. Build the Docker images using Docker Compose:

```bash
docker-compose build
```

2. Run the application:

```bash
docker-compose up
```

The server will be accessible on port `8085`, and the client can be accessed by opening your web browser and navigating to `localhost:8086` (or the appropriate client port).

### Interaction with the Thermal Camera

The server code will upgrade the HTTP connection to a WebSocket connection and listen for incoming commands defined in the `thermalcamera.proto` file. It will also handle commands to adjust the thermal camera's settings, such as zoom level and color scheme.

### Synchronized Views

The new `index.html` includes two iframes that load `ctl.html`, demonstrating synchronized control of the thermal camera. Adjustments made in one view will reflect across all others, thanks to the WebSocket synchronization.

## Additional Information

The server code in Go communicates with the C server using named pipes, defined in `cinterface.go`. This allows for inter-process communication and control over a simulated thermal camera.

Please consult the comments in the `main.go`, `cinterface.go`, and `Dockerfile` for a more detailed understanding of the implementation. If the client includes a GUI, users can send commands to adjust settings and receive responses through the WebSocket connection.

## Stopping the Application

To stop the application, use the following command:

```bash
docker-compose down
```
