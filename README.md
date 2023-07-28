# Thermal Camera API Example

This repository contains a client-server application demonstrating the use of Protobufs and WebSockets to control a thermal camera.

The server is written in Go and the client is a JavaScript-based frontend.

## Repository Structure

- `/server`: Contains the Go server code (`main.go`).
- `/client`: Contains the frontend JavaScript code (`client.js`) and HTML (`index.html`).
- `thermalcamera.proto`: The Protobuf schema defining the service and messages.
- `Dockerfile.client` & `Dockerfile.server`: Dockerfiles used to build the client and server containers.
- `docker-compose.yml`: Docker Compose file to orchestrate the client and server containers.

## Usage

You need Docker and Docker Compose installed on your machine to run the application.

1. Build the Docker images using Docker Compose:

```bash
docker-compose build
```

2. Run the application:

```bash
docker-compose up
```

The server will be accessible on port `8085` and the client can be accessed by opening your web browser and navigating to `localhost:8086`.

The server code will upgrade the HTTP connection to a WebSocket connection and listen for incoming commands defined in the `thermalcamera.proto` file. It will also start streaming the current server time to the client.

The client provides a simple interface to send commands to the server to adjust the thermal camera's settings, such as zoom level and color scheme.

3. To stop the application, use the following command:

```bash
docker-compose down
```
