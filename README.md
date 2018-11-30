# nbody-utils
A collection of routines for reading and efficiently analyzing N-body data.

This library is a cleaned up, documented, and more user-friendly version of the internal utility and I/O libraries used by [Shellfish](https://github.com/phil-mansfield/shellfish) and [gotetra](https://github.com/phil-mansfield/gotetra). Its advatnage is that I've learned a lot about what does and doesn't work, and its interfaces don't need to maintain backwards compatibility with ~50k lines of analysis code.

## Contents:

`array/`

The array package contains utility array functions that allows for easy dataset cutting like in numpy as well as efficient implementations of several sorting and percentile algorithms which can outperform the standard Go library by a factor of x6 - x3.
