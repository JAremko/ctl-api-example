const zoomLevelSelect = document.getElementById('zoom-level');
const colorSchemeSelect = document.getElementById('color-scheme');
const chargeProgressBar = document.getElementById('charge-bar');
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
    const buffer = new Uint8Array(event.data);

    // Attempt to decode as a Command message
    try {
        const commandMessage = Command.decode(buffer);

        if (commandMessage.setZoomLevel) {
            zoomLevelSelect.value = commandMessage.setZoomLevel.level;
        }

        if (commandMessage.setColorScheme) {
            colorSchemeSelect.value = Object.keys(ColorScheme).find(key => ColorScheme[key] === commandMessage.setColorScheme.scheme);
        }
    } catch (error) {
        // If it fails, attempt to decode as a StreamChargeResponse message
        try {
            const response = StreamChargeResponse.decode(buffer);
            const charge = response.charge;
            chargeProgressBar.style.width = `${charge}%`;
            chargePercentage.textContent = `${charge}%`;
        } catch (error) {
            console.error("Failed to decode the message from the server:", error);
        }
    }
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
