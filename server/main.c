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

typedef struct {
  uint32_t id;
  char payload[4];
} Packet;

int main() {
  mkfifo(PIPE_NAME_TO_C, 0600);
  mkfifo(PIPE_NAME_FROM_C, 0600);

  srand(time(NULL));

  int pipeToC = open(PIPE_NAME_TO_C, O_RDONLY);
  int pipeFromC = open(PIPE_NAME_FROM_C, O_WRONLY);

  while (1) {
    Packet packet;
    packet.id = STREAM_CHARGE_RESPONSE;

    int32_t charge = rand() % 101;
    memcpy(packet.payload, &charge, sizeof(charge));
    write(pipeFromC, &packet, sizeof(packet));

    read(pipeToC, &packet, sizeof(packet));

    switch (packet.id) {
    case SET_ZOOM_LEVEL:
      printf("SetZoomLevel command received with level: %d\n", *(int32_t *)packet.payload);
      fflush(stdout);
      break;
    case SET_COLOR_SCHEME:
      printf("SetColorScheme command received with scheme: %d\n", *(int32_t *)packet.payload);
      fflush(stdout);
      break;
    }

    sleep(1);
  }

  close(pipeFromC);
  close(pipeToC);

  return 0;
}
