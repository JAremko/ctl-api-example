# Table of Contents
- [Thermal Camera API Example](#thermal-camera-api-example)
  - [Introduction](#introduction)
  - [Repository Structure](#repository-structure)
  - [Protobuf Definitions (`thermalcamera.proto`)](#protobuf-definitions-thermalcameraproto)
  - [Usage](#usage)
    - [Building and Running the Application](#building-and-running-the-application)
      - [Using Docker Compose (Default)](#using-docker-compose-default)
      - [Using Host for C Server (Alternative)](#using-host-for-c-server-alternative)
    - [Interaction with the Thermal Camera](#interaction-with-the-thermal-camera)
    - [Synchronized Views](#synchronized-views)
  - [Additional Information](#additional-information)
  - [Stopping the Application](#stopping-the-application)

# Thermal Camera API Example

## Introduction

This repository contains a client-server application demonstrating the use of Protobufs and WebSockets to control a thermal camera. The server part consists of two separate components: a Go server and a C server, communicating via named pipes. The client is a JavaScript-based frontend that includes synchronization between clients.

## Repository Structure

- `/server`: Contains the Go server and C server code.
- `/client`: Contains the frontend JavaScript code (`client.js`) and HTML (`ctl.html` and `index.html`).
- `thermalcamera.proto`: The Protobuf schema defining messages.
- `Dockerfile.client`, `Dockerfile.go-server`, `Dockerfile.c-server`: Dockerfiles used to build the client, Go server, and C server containers respectively.
- `docker-compose.yml` and `docker-compose-host.yml`: Docker Compose files to orchestrate the client and server containers.

## Protobuf Definitions (`thermalcamera.proto`)

The `thermalcamera.proto` file defines the messages and services for controlling a thermal camera. The definitions include:

- **Payload**: A wrapper message that includes one of the following payload types:
  - **SetZoomLevel**: To set the zoom level of the camera.
  - **SetColorScheme**: To set the color scheme of the camera.
  - **AccChargeLevel**: To get the accumulated charge level of the camera, represented as a percentage (constantly streamed from the C side).
- **ColorScheme**: An enumeration of available color schemes, including UNKNOWN, SEPIA, BLACK_HOT, and WHITE_HOT.

These definitions are used to serialize the data sent between the client and server, ensuring a consistent and robust communication protocol.

## Usage

### Building and Running the Application

#### Using Docker Compose (Default)

1. Build the Docker images using Docker Compose:

   ```bash
   docker-compose build
   ```

2. Run the application:

   ```bash
   docker-compose up
   ```

#### Using Host for C Server (Alternative)

1. Build the C server on the host:

   ```bash
   gcc -o c_server server/main.c -lpthread
   chmod +x c_server
   ```

2. Create named pipes on the host:

   ```bash
   mkfifo /tmp/toC
   mkfifo /tmp/fromC
   ```

3. Run the C server on the host:

   ```bash
   ./c_server
   ```

4. Build the Docker images using Docker Compose with the alternative file:

   ```bash
   docker-compose -f docker-compose-host.yml build
   ```

5. Run the application:

   ```bash
   docker-compose -f docker-compose-host.yml up
   ```

The Go server will be accessible on port `8085`, and the client can be accessed by opening your web browser and navigating to `localhost:8086`.

### Interaction with the Thermal Camera

The Go server code will upgrade the HTTP connection to a WebSocket connection and listen for incoming commands defined in the `thermalcamera.proto` file. It will also handle commands to adjust the thermal camera's settings, such as zoom level and color scheme.

### Synchronized Views

`index.html` includes two iframes that load `ctl.html`, demonstrating synchronized control of the thermal camera. Adjustments made in one view will reflect across all others.

## Additional Information

The server code in Go communicates with the C server using named pipes, defined in `cinterface.go`. This allows for inter-process communication and control over a simulated thermal camera. The named pipes `/tmp/toC` and `/tmp/fromC` are shared between the Go and C server containers through a shared volume.

## Stopping the Application

To stop the application, use the following command:

```bash
docker-compose down -v
```

Or, if using the host configuration:

```bash
docker-compose -f docker-compose-host.yml down -v
```
