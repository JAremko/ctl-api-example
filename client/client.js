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

let Payload;
let SetZoomLevel;
let SetColorScheme;
let AccChargeLevel;

protobuf.load("thermalcamera.proto").then(root => {
    Payload = root.lookup("thermalcamera.Payload");
    SetZoomLevel = root.lookup("thermalcamera.SetZoomLevel");
    SetColorScheme = root.lookup("thermalcamera.SetColorScheme");
    AccChargeLevel = root.lookup("thermalcamera.AccChargeLevel");
}).catch(error => console.error("Failed to load protobuf definitions:", error));

const handlers = {
    setZoomLevel: payload => {
        zoomLevelSelect.value = payload.setZoomLevel.level;
    },
    setColorScheme: payload => {
        colorSchemeSelect.value = Object.keys(ColorScheme).find(key => ColorScheme[key] === payload.setColorScheme.scheme);
    },
    accChargeLevel: payload => {
        const charge = payload.accChargeLevel.charge;
        chargeProgressBar.style.width = `${charge}%`;
        chargePercentage.textContent = `${charge}%`;
    }
};

const socket = new WebSocket("ws://localhost:8085");
socket.binaryType = "arraybuffer";

socket.onmessage = event => {
    const buffer = new Uint8Array(event.data);
    const payloadMessage = Payload.decode(buffer);

    // Call handlers based on fields present in the payload
    if (payloadMessage.setZoomLevel) handlers.setZoomLevel(payloadMessage);
    if (payloadMessage.setColorScheme) handlers.setColorScheme(payloadMessage);
    if (payloadMessage.accChargeLevel) handlers.accChargeLevel(payloadMessage);
};

zoomLevelSelect.addEventListener('change', () => {
    const setZoomLevelMessage = SetZoomLevel.create({ level: parseInt(zoomLevelSelect.value) });
    const payload = Payload.create({ setZoomLevel: setZoomLevelMessage });
    const buffer = Payload.encode(payload).finish();
    socket.send(buffer);
});

colorSchemeSelect.addEventListener('change', () => {
    const setColorSchemeMessage = SetColorScheme.create({ scheme: ColorScheme[colorSchemeSelect.value] });
    const payload = Payload.create({ setColorScheme: setColorSchemeMessage });
    const buffer = Payload.encode(payload).finish();
    socket.send(buffer);
});
