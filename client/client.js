const zoomLevelSelect = document.getElementById('zoom-level');
const colorSchemeSelect = document.getElementById('color-scheme');
const chargeProgressBar = document.getElementById('charge-progress');
const chargePercentage = document.getElementById('charge-percentage');

const ColorScheme = {
    "UNKNOWN": 0,
    "SEPIA": 1,
    "BLACK_HOT": 2,
    "WHITE_HOT": 3
};

let Command;
let SetZoomLevel;
let SetColorScheme;
let StreamChargeResponse;

protobuf.load("thermalcamera.proto").then(root => {
    Command = root.lookup("thermalcamera.Command");
    SetZoomLevel = root.lookup("thermalcamera.SetZoomLevel");
    SetColorScheme = root.lookup("thermalcamera.SetColorScheme");
    StreamChargeResponse = root.lookup("thermalcamera.StreamChargeResponse");
}).catch(error => console.error("Failed to load protobuf definitions:", error));

const socket = new WebSocket("ws://localhost:8085");
socket.binaryType = "arraybuffer";

socket.onmessage = event => {
    const response = StreamChargeResponse.decode(new Uint8Array(event.data));
    const charge = response.charge;
    const chargeBar = document.getElementById('charge-bar');
    chargeBar.style.width = `${charge}%`;
    chargePercentage.textContent = `${charge}%`;
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
