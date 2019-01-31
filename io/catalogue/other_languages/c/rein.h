#ifndef __REIN_H
#define __REIN_H

#include <stdint.h>
#include <stdio.h>

typedef enum {
    rein_int64_t,
    rein_uint64_t,
    rein_int32_t,
    rein_uint32_t,
    rein_int16_t,
    rein_uint16_t,
    rein_int8_t,
    rein_uint8_t,
    rein_float32_t,
    rein_double_t
} rein_type;

typedef uint64_t splitmix64_state;
typedef uint64_t column_flag;

typedef struct {
    FILE *fp;
    splitmix64_state rand_state;

    // Fixed width header values

    uint64_t version;
    int64_t seed, columns, mass_column, blocks, text_header_length;
    int64_t text_column_names_length, is_sorted;
    double min_mass;

    // Array header values

    double *deltas;
    uint8_t *column_skipped;
    char *text_header, *text_column_names;

    // block values

    int64_t *block_haloes;
    column_flag **block_flags;
    int64_t **block_keys;
    size_t *data_offsets;
    
} rein_binh;

rein_binh *rein_binh_open(char *fname);
void rein_binh_close(rein_binh *file);

void rein_binh_read_column_block(
    rein_binh *file, int block,
    int column,
    rein_type t, void *buffer
);
void rein_binh_read_column(
    rein_binh *file, int block,
    rein_type t, void *buffer
);

int rein_binh_column_index(rein_binh *file, char *name);

#endif
