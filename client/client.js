// DOM elements for user interaction
const zoomLevelSelect = document.getElementById('zoom-level');
const colorSchemeSelect = document.getElementById('color-scheme');
const chargeProgressBar = document.getElementById('charge-bar');
const chargePercentage = document.getElementById('charge-percentage');

// Enumeration for color schemes
const ColorScheme = {
    "UNKNOWN": 0,
    "SEPIA": 1,
    "BLACK_HOT": 2,
    "WHITE_HOT": 3
};

// Protobuf message types will be loaded here
let Payload;
let SetZoomLevel;
let SetColorScheme;
let AccChargeLevel;

// Load the Protobuf definitions from the file
protobuf.load("thermalcamera.proto").then(root => {
    // Initialize the message types
    Payload = root.lookup("thermalcamera.Payload");
    SetZoomLevel = root.lookup("thermalcamera.SetZoomLevel");
    SetColorScheme = root.lookup("thermalcamera.SetColorScheme");
    AccChargeLevel = root.lookup("thermalcamera.AccChargeLevel");

    // If you want to add more message types, follow the pattern above to define them
}).catch(error => console.error("Failed to load protobuf definitions:", error));

// Define handlers for different payload types
const handlers = {
    setZoomLevel: payload => {
        // Update the zoom level on UI
        zoomLevelSelect.value = payload.setZoomLevel.level;
    },
    setColorScheme: payload => {
        // Update the color scheme on UI
        colorSchemeSelect.value = Object.keys(ColorScheme).find(key => ColorScheme[key] === payload.setColorScheme.scheme);
    },
    accChargeLevel: payload => {
        // Update the charge progress bar and percentage on UI
        const charge = payload.accChargeLevel.charge;
        chargeProgressBar.style.width = `${charge}%`;
        chargePercentage.textContent = `${charge}%`;
    },
    // Add more handlers here for new payload types
};

// Create a WebSocket connection to the server
const socket = new WebSocket("ws://localhost:8085");
socket.binaryType = "arraybuffer";

// Handle incoming messages from the server
socket.onmessage = event => {
    // Decode the payload message using Protobuf
    const buffer = new Uint8Array(event.data);
    const payloadMessage = Payload.decode(buffer);

    // Call the appropriate handler based on fields present in the payload
    if (payloadMessage.setZoomLevel) handlers.setZoomLevel(payloadMessage);
    if (payloadMessage.setColorScheme) handlers.setColorScheme(payloadMessage);
    if (payloadMessage.accChargeLevel) handlers.accChargeLevel(payloadMessage);

    // Add more conditional checks here to call new handlers for new payload types
};

// Event listener for zoom level changes
zoomLevelSelect.addEventListener('change', () => {
    // Create and send a message to set zoom level
    const setZoomLevelMessage = SetZoomLevel.create({ level: parseInt(zoomLevelSelect.value) });
    const payload = Payload.create({ setZoomLevel: setZoomLevelMessage });
    const buffer = Payload.encode(payload).finish();
    socket.send(buffer);
});

// Event listener for color scheme changes
colorSchemeSelect.addEventListener('change', () => {
    // Create and send a message to set color scheme
    const setColorSchemeMessage = SetColorScheme.create({ scheme: ColorScheme[colorSchemeSelect.value] });
    const payload = Payload.create({ setColorScheme: setColorSchemeMessage });
    const buffer = Payload.encode(payload).finish();
    socket.send(buffer);
});

// To add more types of messages:
// 1. Define new Protobuf message types in "thermalcamera.proto"
// 2. Load and initialize them like other message types above
// 3. Create handlers for them in the "handlers" object
// 4. Add conditional checks in the "socket.onmessage" handler
// 5. If needed, add new UI elements and event listeners to trigger these messages
