# nbody-utils
A collection of routines for reading and efficiently analyzing cosmological N-body data.

This library is a cleaned up, documented, and more user-friendly version of the internal utility and I/O libraries used by [Shellfish](https://github.com/phil-mansfield/shellfish) and [gotetra](https://github.com/phil-mansfield/gotetra). Its advatnage is that I've learned a lot about what does and doesn't work, and its interfaces don't need to maintain backwards compatibility with ~50k lines of analysis code.

## Contents:

`array/`

The array package contains utility array functions that allows for easy dataset cutting like in numpy as well as efficient implementations of several sorting and percentile algorithms which can outperform the standard Go library by a factor of x6 - x3.

For example, suppose you have x, y, z, vmax, and pid halo data from a catalog, want to cut your sample by 100 < vmax < 200, remove subhaloes with pid >= 0, and then sort by vmax in descending order. This is a pain to do with normal for loops, but looks like this with the array library:

```go
import (
    ar "github.com/phil-mansifle/nbody-utils/array"
)

func main() {
    x, y, z, vmax, pid := // Read from file
    
    // Cut data
    ok := ar.And(ar.Greater(vmax, 100), ar.Less(vmax, 100), ar.IntGeq(pid, 0))
    x, y, z = ar.Cut(x, ok), ar.Cut(y, ok), ar.Cut(z, ok)
    vmax, pid = ar.Cut(vmax, ok), ar.IntCut(pid, ok)
    
    // Sort data
    order := ar.IntReverse(ar.QuickSortIndex(vmax))
    x, y, z = ar.Order(x, order), ar.Order(y, order), ar.Order(z, order)
    vmax, pid = ar.Order(vmax, order), ar.IntOrder(pid, order)
    
    // Do analysis
}
```
