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

typedef struct {
  uint32_t id;
  char payload[PayloadSize];
} Packet;

void *handleCommands(void *args) {
  int pipeToC = open(PIPE_NAME_TO_C, O_RDONLY);

  while (1) {
    Packet packet;
    read(pipeToC, &packet, sizeof(packet));
    switch (packet.id) {
    case SET_ZOOM_LEVEL:
      printf("[C] SetZoomLevel command received with level: %d\n", *(int32_t *)packet.payload);
      break;
    case SET_COLOR_SCHEME:
      printf("[C] SetColorScheme command received with scheme: %d\n", *(int32_t *)packet.payload);
      break;
    }
    fflush(stdout);
  }

  close(pipeToC);
  return NULL;
}

void *updateCharge(void *args) {
  int pipeFromC = open(PIPE_NAME_FROM_C, O_WRONLY);

  srand(time(NULL));

  while (1) {
    Packet packet;
    packet.id = STREAM_CHARGE_RESPONSE;

    int32_t charge = rand() % 101;
    memcpy(packet.payload, &charge, sizeof(charge));
    write(pipeFromC, &packet, sizeof(packet));

    sleep(1);
  }

  close(pipeFromC);
  return NULL;
}

int main() {
  mkfifo(PIPE_NAME_TO_C, 0600);
  mkfifo(PIPE_NAME_FROM_C, 0600);

  pthread_t commandThread, chargeThread;

  pthread_create(&commandThread, NULL, handleCommands, NULL);
  pthread_create(&chargeThread, NULL, updateCharge, NULL);

  pthread_join(commandThread, NULL);
  pthread_join(chargeThread, NULL);

  return 0;
}
