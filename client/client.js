const zoomLevelSelect = document.getElementById('zoom-level');
const colorSchemeSelect = document.getElementById('color-scheme');
const serverTimeInput = document.getElementById('server-time');

let Command;
let SetZoomLevelRequest;
let SetColorSchemeRequest;
let StreamTimeResponse;

protobuf.load("thermalcamera.proto").then(root => {
    Command = root.lookup("thermalcamera.Command");
    SetZoomLevelRequest = root.lookup("thermalcamera.SetZoomLevelRequest");
    SetColorSchemeRequest = root.lookup("thermalcamera.SetColorSchemeRequest");
    StreamTimeResponse = root.lookup("thermalcamera.StreamTimeResponse");
}).catch(error => console.error("Failed to load protobuf definitions:", error));

const socket = new WebSocket("ws://localhost:8085");
socket.binaryType = "arraybuffer";

socket.onmessage = event => {
    const response = StreamTimeResponse.decode(new Uint8Array(event.data));
    serverTimeInput.value = response.time;
};

zoomLevelSelect.onchange = () => {
    try {
        console.log("Creating SetZoomLevel message");
        const setZoomLevelMessage = SetZoomLevelRequest.create({ level: parseInt(zoomLevelSelect.value) });
        console.log("Creating Command message");
        const commandZoomLevel = Command.create({ setZoomLevel: setZoomLevelMessage });
        console.log("Encoding Command message");
        const commandZoomLevelBuffer = Command.encode(commandZoomLevel).finish();
        console.log("Sending message");
        socket.send(commandZoomLevelBuffer);
    } catch (error) {
        console.error("Error during zoomLevelSelect onchange:", error);
    }
};

colorSchemeSelect.onchange = () => {
    try {
        console.log("Creating SetColorScheme message");
        const setColorSchemeMessage = SetColorSchemeRequest.create({ scheme: colorSchemeSelect.value });
        console.log("Creating Command message");
        const commandColorScheme = Command.create({ setColorScheme: setColorSchemeMessage });
        console.log("Encoding Command message");
        const commandColorSchemeBuffer = Command.encode(commandColorScheme).finish();
        console.log("Sending message");
        socket.send(commandColorSchemeBuffer);
    } catch (error) {
        console.error("Error during colorSchemeSelect onchange:", error);
    }
};
