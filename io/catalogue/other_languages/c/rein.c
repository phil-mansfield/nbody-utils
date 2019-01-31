#include "rein.h"
#include <stdlib.h>
#include <stdbool.h>
#include <inttypes.h>

// Byteswapping macros from glibc's byteswap.h
#define bswap_16(x)                                             \
    ((__uint16_t) ((((x) >> 8) & 0xff) | (((x) & 0xff) << 8)))

#define bswap_32(x)                                                 \
    ((((x) & 0xff000000u) >> 24) | (((x) & 0x00ff0000u) >> 8)       \
     | (((x) & 0x0000ff00u) << 8) | (((x) & 0x000000ffu) << 24))

#define bswap_64(x)                             \
    ((((x) & 0xff00000000000000ull) >> 56)      \
     | (((x) & 0x00ff000000000000ull) >> 40)    \
     | (((x) & 0x0000ff0000000000ull) >> 24)    \
     | (((x) & 0x000000ff00000000ull) >> 8)     \
     | (((x) & 0x00000000ff000000ull) << 8)     \
     | (((x) & 0x0000000000ff0000ull) << 24)    \
     | (((x) & 0x000000000000ff00ull) << 40)    \
     | (((x) & 0x00000000000000ffull) << 56))

bool is_little_endian() {
    uint16_t x = 1;
    return *(uint8_t*)(&x) == 1;
}

/* splitmix64 was written by Sebastiano Vigna and its original source code
 * can be found here: http://xoshiro.di.unimi.it/splitmix64.c */

uint64_t splitmix64_uint64(splitmix64_state *state) {
    *state += 0x9e3779b97f4a7c15;
    uint64_t z = *state;
    z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9;
    z = (z ^ (z >> 27)) * 0x94d049bb133111eb;
    return z ^ (z >> 31);
}

double splitmix64_double(splitmix64_state *state) {
    uint64_t x = splitmix64_uint64(state);
    return x / (double) UINT64_MAX;
}

float splitmix64_float(splitmix64_state *state) {
    return (float) splitmix64_double(state);
}

/*bswap(bytes, size, len) swaps the endianness of a byte buffer containing
 * `len` elements each with `size` bytes. Only acceptable sizes are 8, 4,
 * and 2. */
void bswap(void *bytes, size_t size, size_t len) {
    if (size == 8) {
        uint64_t *data = (uint64_t*)bytes;
        for (size_t i = 0; i < len; i++) {
            uint64_t x = data[i];
            data[i] = bswap_64(x);
        }
    } else if (size == 4) {
        uint32_t *data = (uint32_t*)bytes;
        for (size_t i = 0; i < len; i++) {
            uint32_t x = data[i];
            data[i] = bswap_32(x);
        }
    } else if (size == 2) {
        uint16_t *data = (uint16_t*)bytes;
        for (size_t i = 0; i < len; i++) {
            uint16_t x = data[i];
            data[i] = bswap_16(x);
        }
    } else {
        fprintf(stderr, "Unknown element size %zu.\n", size);
        exit(1);
    }
}

void LE_read(void *buf, size_t size, size_t len, FILE *fp) {
    size_t bytes_read = fread(buf, size, len, fp);
    if (bytes_read != size*len) {
        fprintf(stderr, "fread only read %zu bytes, even thought size = %zu " 
                "and len = %zu.\n", bytes_read, size, len);
        exit(1);
    }
    
    if (!is_little_endian()) {
        bswap(buf, size, len);
    }
}

size_t next_block_start(rein_binh* file, int64_t block, size_t block_start) {
    (void) file;
    (void) block;
    (void) block_start;
    return 0;
}

rein_binh *rein_binh_open(char *fname) {
    // Intialize

    rein_binh *file = calloc(1, sizeof(*file));
    file->fp = fopen(fname, "rb");
    if (file->fp == NULL) {
        fprintf(stderr, "Could not open file %s.\n", fname);
        exit(1);
    }

    // Check the version

    LE_read(&file->version, 8, 1, file->fp);
    if(file->version != 2) {
        fprintf(stderr, "file %s uses version %"PRIu64", but I/O library is "
                "version 2.", fname, file->version);
    }

    // Read all the easy fields

    LE_read(&file->seed, 8, 8, file->fp);
    file->deltas = calloc((size_t)file->columns, sizeof(double));
    file->column_skipped = calloc((size_t)file->columns, sizeof(uint8_t));
    file->text_header = calloc((size_t)file->text_header_length+1, 1);
    file->text_column_names = calloc(
        (size_t)file->text_column_names_length+1, 1
    );
    LE_read(&file->deltas, 8, (size_t)file->columns, file->fp);
    LE_read(&file->column_skipped, 1, (size_t)file->columns, file->fp);
    LE_read(&file->text_header, 1, (size_t)file->text_header_length, file->fp);
    LE_read(
        &file->text_column_names, 1,
        (size_t)file->text_column_names_length, file->fp
    );

    // Set up RNG

    file->rand_state = (splitmix64_state)file->seed;

    // Now it's time for the block-based fields

    size_t header_size = 8*9 + (8 + 1)*(size_t)file->columns +
        (size_t)file->text_header_length +
        (size_t)file->text_column_names_length;

    size_t block_start = header_size;
    file->block_haloes = calloc((size_t)file->columns, sizeof(int64_t));
    file->block_flags = calloc((size_t)file->columns, sizeof(column_flag*));
    file->block_keys = calloc((size_t)file->columns, sizeof(int64_t*));
    file->data_offsets = calloc((size_t)file->columns, sizeof(size_t));

    for (int64_t block = 0; block < file->blocks; block++) {
        file->data_offsets[block] =
            block_start + 8*(1 + 2*(size_t)file->columns);
        _fseeki64(file->fp, (int64_t)block_start, SEEK_SET);

        LE_read(file->block_healos[block], 8, 1, file->fp);
        file->block_flags[block] = calloc(
            (size_t)file->columns, sizeof(block_flag)
        );
        LE_read(file->block_keys[block], 8, (size_t)file->columns, file->fp);
        file->block_keys[block] = calloc(
            (size_t)file->columns, sizeof(int64_t)
        );

        block_start = next_block_header(file, block, block_start);
    }
    
    return NULL;
}

void rein_binh_close(rein_binh *file) {
    (void) file;

    return;
}

void rein_binh_read_column_block(
    rein_binh *file, int block,
    int column,
    rein_type t, void *buffer
) {
    (void) file;
    (void) block;
    (void) column;
    (void) t;
    (void) buffer;

    return;
}

void rein_binh_read_column(
    rein_binh *file, int column,
    rein_type t, void *buffer
) {
    (void) file;
    (void) column;
    (void) t;
    (void) buffer;

    return;
}

int rein_binh_column_index(rein_binh *file, char *name) {
    (void) file;
    (void) name;

    return 0;
}

int main() {
    return 0;
}
