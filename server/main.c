#include <pthread.h>
#include <fcntl.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <time.h>
#include <unistd.h>

#define PIPE_NAME_TO_C "/tmp/toC"
#define PIPE_NAME_FROM_C "/tmp/fromC"
#define SET_ZOOM_LEVEL 1
#define SET_COLOR_SCHEME 2
#define STREAM_CHARGE_RESPONSE 3
#define PayloadSize 64

// Packet structure for communication between processes
typedef struct {
  uint32_t id;
  char payload[PayloadSize];
} Packet;

// Function to handle incoming commands from another process
void *handleCommands(void *args) {
  // Open named pipe for reading
  int pipeToC = open(PIPE_NAME_TO_C, O_RDONLY);
  if (pipeToC < 0) {
    perror("[C] Error opening pipe to C");
    return NULL;
  }

  while (1) {
    Packet packet;
    // Read the incoming packet
    ssize_t bytesRead = read(pipeToC, &packet, sizeof(packet));
    if (bytesRead < 0) {
      perror("[C] Error reading from pipe");
      break;
    }

    // Handle the packet based on its ID
    switch (packet.id) {
    case SET_ZOOM_LEVEL:
      printf("[C] SetZoomLevel command received with level: %d\n", *(int32_t *)packet.payload);
      fflush(stdout);
      break;
    case SET_COLOR_SCHEME:
      printf("[C] SetColorScheme command received with scheme: %d\n", *(int32_t *)packet.payload);
      fflush(stdout);
      break;
    }
  }

  close(pipeToC);
  return NULL;
}

// Function to update the charge and send it to Go process
void *updateCharge(void *args) {
  // Open named pipe for writing
  int pipeFromC = open(PIPE_NAME_FROM_C, O_WRONLY);
  if (pipeFromC < 0) {
    perror("[C] Error opening pipe from C");
    return NULL;
  }

  srand(time(NULL));

  while (1) {
    Packet packet;
    packet.id = STREAM_CHARGE_RESPONSE;

    int32_t charge = rand() % 101;
    memcpy(packet.payload, &charge, sizeof(charge));
    // Write the packet to the pipe
    if (write(pipeFromC, &packet, sizeof(packet)) < 0) {
      perror("[C] Error writing to pipe");
      break;
    }

    sleep(1);
  }

  close(pipeFromC);
  return NULL;
}

int main() {
  // Remove existing named pipes if they exist
  unlink(PIPE_NAME_TO_C);
  unlink(PIPE_NAME_FROM_C);

  // Create named pipes
  mkfifo(PIPE_NAME_TO_C, 0600);
  mkfifo(PIPE_NAME_FROM_C, 0600);

  pthread_t commandThread, chargeThread;

  // Create threads to handle commands and update charge
  pthread_create(&commandThread, NULL, handleCommands, NULL);
  pthread_create(&chargeThread, NULL, updateCharge, NULL);

  // Wait for the threads to complete
  pthread_join(commandThread, NULL);
  pthread_join(chargeThread, NULL);

  // Remove the named pipes
  unlink(PIPE_NAME_TO_C);
  unlink(PIPE_NAME_FROM_C);

  return 0;
}
