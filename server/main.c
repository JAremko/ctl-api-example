#include <pthread.h>
#include <fcntl.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <time.h>
#include <unistd.h>

#define PIPE_NAME_TO_C "/tmp/toC"       // Named pipe for communication TO C process
#define PIPE_NAME_FROM_C "/tmp/fromC"   // Named pipe for communication FROM C process
#define SET_ZOOM_LEVEL 1                // Command ID for setting zoom level
#define SET_COLOR_SCHEME 2              // Command ID for setting color scheme
#define CHARGE_PACKET 3                 // Packet ID for sending charge level
#define PayloadSize 64                  // Maximum size of packet payload

typedef struct {
  uint32_t id;                         // Identifies the type of packet (command or charge)
  char payload[PayloadSize];           // Contains data for commands or state updates
} Packet;

pthread_mutex_t pipe_mutex = PTHREAD_MUTEX_INITIALIZER; // Mutex for synchronizing pipe access among threads

// Handler for zoom level command
void handleZoomLevel(Packet *packet) {
  int32_t zoomLevel = *(int32_t *)packet->payload; // Extract the zoom level from the payload
  printf("[C] SetZoomLevel command received with level: %d\n", zoomLevel);
  *(int32_t *)packet->payload = zoomLevel; // Identity modification; Assigning the same value back as a dummy operation
}

// Handler for color scheme command
void handleColorScheme(Packet *packet) {
  int32_t colorScheme = *(int32_t *)packet->payload; // Extract the color scheme from the payload
  printf("[C] SetColorScheme command received with scheme: %d\n", colorScheme);
  *(int32_t *)packet->payload = colorScheme; // Identity modification; Assigning the same value back as a dummy operation
}

// Function to process incoming commands, update the internal state of the device, and log information
void processCommand(Packet *packet) {
  switch (packet->id) {
  case SET_ZOOM_LEVEL:
    handleZoomLevel(packet);           // Call handler for zoom level
    break;
  case SET_COLOR_SCHEME:
    handleColorScheme(packet);         // Call handler for color scheme
    break;
  }
  fflush(stdout);                      // Flush the stdout buffer to print log messages immediately
}

void *handleCommands(void *args) {
  int pipeToC = open(PIPE_NAME_TO_C, O_RDONLY);      // Open the pipe for reading commands from another process
  int pipeFromC = open(PIPE_NAME_FROM_C, O_WRONLY);  // Open the pipe for writing responses to another process

  if (pipeToC < 0 || pipeFromC < 0) {
    perror("[C] Error opening pipes");
    return NULL;
  }

  while (1) {
    Packet packet;
    ssize_t bytesRead = read(pipeToC, &packet, sizeof(packet)); // Read command packet from pipe
    if (bytesRead < 0) {
      perror("[C] Error reading from pipe");
      break;
    }

    processCommand(&packet);          // Process the command using appropriate handler

    pthread_mutex_lock(&pipe_mutex);  // Lock the mutex to ensure exclusive access to the pipe
    if (write(pipeFromC, &packet, sizeof(packet)) < 0) { // Write the response packet back to the pipe
      perror("[C] Error writing to pipe");
      pthread_mutex_unlock(&pipe_mutex);
      break;
    }
    pthread_mutex_unlock(&pipe_mutex); // Unlock the mutex to allow other threads to access the pipe
  }

  close(pipeToC);                     // Close the read pipe
  close(pipeFromC);                   // Close the write pipe
  return NULL;
}

void *updateCharge(void *args) {
  int pipeFromC = open(PIPE_NAME_FROM_C, O_WRONLY);  // Open the pipe for writing charge updates to another process

  if (pipeFromC < 0) {
    perror("[C] Error opening pipe from C");
    return NULL;
  }

  srand(time(NULL));                   // Seed random number generator for simulating charge values

  while (1) {
    Packet packet;
    packet.id = CHARGE_PACKET;

    int32_t charge = rand() % 101;     // Generate a random charge value between 0 and 100
    memcpy(packet.payload, &charge, sizeof(charge)); // Copy the charge value to the payload

    pthread_mutex_lock(&pipe_mutex);   // Lock the mutex to ensure exclusive access to the pipe
    if (write(pipeFromC, &packet, sizeof(packet)) < 0) { // Write the charge packet to the pipe
      perror("[C] Error writing to pipe");
      pthread_mutex_unlock(&pipe_mutex);
      break;
    }
    pthread_mutex_unlock(&pipe_mutex); // Unlock the mutex to allow other threads to access the pipe

    sleep(1);                          // Wait for 1 second before generating the next charge update
  }

  close(pipeFromC);                    // Close the write pipe
  return NULL;
}

int main() {
  unlink(PIPE_NAME_TO_C);              // Remove any existing named pipe TO C process
  unlink(PIPE_NAME_FROM_C);            // Remove any existing named pipe FROM C process

  mkfifo(PIPE_NAME_TO_C, 0600);        // Create named pipe TO C process with read/write permissions for the owner
  mkfifo(PIPE_NAME_FROM_C, 0600);      // Create named pipe FROM C process with read/write permissions for the owner

  pthread_t commandThread, chargeThread;
  pthread_create(&commandThread, NULL, handleCommands, NULL); // Create a thread to handle incoming commands
  pthread_create(&chargeThread, NULL, updateCharge, NULL);    // Create a thread to periodically update the charge

  pthread_join(commandThread, NULL);   // Wait for the command handling thread to terminate
  pthread_join(chargeThread, NULL);    // Wait for the charge update thread to terminate

  unlink(PIPE_NAME_TO_C);              // Remove the named pipe TO C process
  unlink(PIPE_NAME_FROM_C);            // Remove the named pipe FROM C process

  return 0;
}
