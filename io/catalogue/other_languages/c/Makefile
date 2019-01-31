LARGEFILE_FLAGS=-D_FILE_OFFSET_BITS=64 -D_LARGEFILE_SOURCE -D_LARGEFILE64_SOURCE

CLANG_CFLAGS=-Werror -Weverything -Wno-unused-macros -Wno-missing-prototypes -Wno-bad-function-cast -Wno-float-equal -Wno-padded -Wno-missing-noreturn -std=c99 -pedantic -g -O2 

GCC_CFLAGS= -O2 -Wall -Wextra -Werror -std=c99 -pedantic -Wshadow -Wcast-qual -Wcast-align -Wundef -Wredundant-decls -Wmissing-include-dirs -g

# icc generally generates too many remarks (small warnings), some of which are
# caused by very benign actions which are impossible to avoid. If you want to
# remove more remarks, you can add to the comma-separated list at the end of
# the flag set.
#
# Currently removed icc remarks:
# 1419 - Disallows forward declarations in source files.
# 1572 - Floating point equality and inequalities.
# 2259 - This is a well meaning warning which warns out non-explicit casts,
#        but it also watns about, e.g., subtracting ints.
# 2415 - Unused static variable declarations.

ICC_CFLAGS= -O2 -g -Werror -std=c99 -pedantic -w3 -wd1419,1572,2259,2415
# Select the compiler and compiler flags that you want to use.
#CC=gcc
#CFLAGS=$(GCC_CFLAGS)
CC=clang
CFLAGS=$(CLANG_CFLAGS)
#CC=icc
#CFLAGS=$(ICC_CFLAGS)

# The debugging flags -D DEBUG_MOCK_BIG_ENDIAN and -D DEBUG_MOCK_LITTLE_ENDIAN
# can be used to force the rpoject to read/write files as if it had a big/little
# endian arcitechture.

# Add/remove any additional files that aren't in src at the end of this list.
SOURCES=$(wildcard *.c)
OBJECTS=$(patsubst %.c,%.o,$(SOURCES))
HEADERS=$(patsubst %.c,%.h,$(SOURCES))

# Location of libraries being used.
LIBRARIES=
# Flags of libraries being used.
LIBRARY_FLAGS=-lm
# Location of .h files which should be included.
INCLUDES=

ifneq ($(INCLUDES),)
	INCLUDES_WITH_FLAG=-I $(INCLUDES)
else
	INCLUDES_WITH_FLAG=$(INCLUDES)
endif

%.o: %.c %.h Makefile
	$(CC) $(CFLAGS) -c -o $@ $< $(INCLUDES_WITH_FLAG)

rein.o: rein.c rein.h

all: $(OBJECTS) $(HEADERS)
	$(CC) $(CFLAGS) rein.o -o rein_test

clean:
	rm *.o rein_test
