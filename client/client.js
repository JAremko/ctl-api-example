const zoomLevelSelect = document.getElementById('zoom-level');
const colorSchemeSelect = document.getElementById('color-scheme');
const serverTimeInput = document.getElementById('server-time');

const ColorScheme = {
    "UNKNOWN": 0,
    "RED": 1,
    "GREEN": 2,
    "BLUE": 3,
    "GRAYSCALE": 4
};

let Command;
let SetZoomLevel;
let SetColorScheme;
let StreamTimeResponse;

protobuf.load("thermalcamera.proto").then(root => {
    Command = root.lookup("thermalcamera.Command");
    SetZoomLevel = root.lookup("thermalcamera.SetZoomLevel");
    SetColorScheme = root.lookup("thermalcamera.SetColorScheme");
    StreamTimeResponse = root.lookup("thermalcamera.StreamTimeResponse");
}).catch(error => console.error("Failed to load protobuf definitions:", error));

const socket = new WebSocket("ws://localhost:8085");
socket.binaryType = "arraybuffer";

socket.onmessage = event => {
    const response = StreamTimeResponse.decode(new Uint8Array(event.data));
    serverTimeInput.value = response.time;
};

zoomLevelSelect.onchange = () => {
    const setZoomLevelMessage = SetZoomLevel.create({ level: parseInt(zoomLevelSelect.value) });
    const commandZoomLevel = Command.create({ setZoomLevel: setZoomLevelMessage });
    const buffer = Command.encode(commandZoomLevel).finish();
    socket.send(buffer);
};

colorSchemeSelect.onchange = () => {
    const setColorSchemeMessage = SetColorScheme.create({ scheme: ColorScheme[colorSchemeSelect.value] });
    const commandColorScheme = Command.create({ setColorScheme: setColorSchemeMessage });
    const buffer = Command.encode(commandColorScheme).finish();
    socket.send(buffer);
};
