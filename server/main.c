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

pthread_mutex_t pipe_mutex = PTHREAD_MUTEX_INITIALIZER;

// Function to handle incoming commands from another process
void *handleCommands(void *args) {
  int pipeToC = open(PIPE_NAME_TO_C, O_RDONLY);
  int pipeFromC = open(PIPE_NAME_FROM_C, O_WRONLY);

  if (pipeToC < 0 || pipeFromC < 0) {
    perror("[C] Error opening pipes");
    return NULL;
  }

  while (1) {
    Packet packet;
    ssize_t bytesRead = read(pipeToC, &packet, sizeof(packet));
    if (bytesRead < 0) {
      perror("[C] Error reading from pipe");
      break;
    }

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

    pthread_mutex_lock(&pipe_mutex);
    if (write(pipeFromC, &packet, sizeof(packet)) < 0) {
      perror("[C] Error writing to pipe");
      pthread_mutex_unlock(&pipe_mutex);
      break;
    }
    pthread_mutex_unlock(&pipe_mutex);
  }

  close(pipeToC);
  close(pipeFromC);
  return NULL;
}

// Function to update the charge and send it to Go process
void *updateCharge(void *args) {
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

    pthread_mutex_lock(&pipe_mutex);
    if (write(pipeFromC, &packet, sizeof(packet)) < 0) {
      perror("[C] Error writing to pipe");
      pthread_mutex_unlock(&pipe_mutex);
      break;
    }
    pthread_mutex_unlock(&pipe_mutex);

    sleep(1);
  }

  close(pipeFromC);
  return NULL;
}

int main() {
  unlink(PIPE_NAME_TO_C);
  unlink(PIPE_NAME_FROM_C);

  mkfifo(PIPE_NAME_TO_C, 0600);
  mkfifo(PIPE_NAME_FROM_C, 0600);

  pthread_t commandThread, chargeThread;
  pthread_create(&commandThread, NULL, handleCommands, NULL);
  pthread_create(&chargeThread, NULL, updateCharge, NULL);

  pthread_join(commandThread, NULL);
  pthread_join(chargeThread, NULL);

  unlink(PIPE_NAME_TO_C);
  unlink(PIPE_NAME_FROM_C);

  return 0;
}
