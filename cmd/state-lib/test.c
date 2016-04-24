#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "libtoxstate.h"

//gcc -Wall -Wl,-rpath,. -L. -l:libtoxstate.so -o test *.c

void print_hex(uint8_t *bytes, size_t length)
{
    for (size_t i = 0; i < length; i++) {
        printf("%02X", bytes[i]);
    }
    printf("\n");
}

int main(int argc, char **argv)
{
    char *file_name = argv[1];
    FILE *file = fopen(file_name, "rb");

	if (file == NULL) {
        printf("could not open %s\n", file_name);
        return 1;
    }

	fseek(file, 0, SEEK_END);
	size_t file_size = ftell(file);
	fseek(file, 0, SEEK_SET);

	uint8_t *data = malloc(file_size * sizeof(uint8_t));
	fread(data, sizeof(uint8_t), file_size, file);
	fclose(file);

    struct Tox_State *state = NULL;
    state_unmarshal(data, file_size, &state);
    free(data);

    if (state == NULL) {
        printf("send help\n");
        return 1;
    }

    printf("nospam: %d\n", state->nospam);
    printf("public_key: ");
    print_hex(state->public_key, sizeof(state->public_key));
    printf("secret_key: ");
    print_hex(state->secret_key, sizeof(state->secret_key));

    printf("name: %s\n", state->name);
    printf("status_message: %s\n", state->status_message);

    for (size_t i = 0; i < state->friends_length; i++) {
        printf("friend name: %s\n", state->friends[i]->name);
    }

    state_free(state);
	return 0;
}
