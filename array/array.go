/*package srray provides functions for sorting and finding the median of
float64 slices without the overhead of Go's interfaces as well as various
array-manipulation utilities.
*/
package array

import (
	"fmt"
)

// Reverse reverses a slice in place (and returns it for convenience).
func Reverse(xs []float64) []float64 {
	n1, n2 := len(xs)-1, len(xs)/2
	for i := 0; i < n2; i++ {
		xs[i], xs[n1-i] = xs[n1-i], xs[i]
	}
	return xs
}

// getOutput is a utility function that gets the output array from an optional
// argument or allocates a new one.
func getOutput(out [][]bool, n int) []bool {
	if len(out) == 0 {
		return make([]bool, n)
	} else {
		ok := out[0]
		if len(ok) != n {
			panic(fmt.Sprintf(
				"len(xs) = %d, but len(out) = %d", n, len(ok)),
			)
		}
		return ok
	}
}

// Greater returns a bool array representing which elements of xs are greater
// than x0. It takes a output target as an optional argument to avoid excess
// allocations.
func Greater(xs []float64, x0 float64, out ...[]bool) []bool {
	ok := getOutput(out, len(xs))
	for i := range xs {
		ok[i] = xs[i] > x0
	}
	return ok
}

// Less returns a bool array representing which elements of xs are less
// than x0. It takes a output target as an optional argument to avoid excess
// allocations.
func Less(xs []float64, x0 float64, out ...[]bool) []bool {
	ok := getOutput(out, len(xs))
	for i := range xs {
		ok[i] = xs[i] < x0
	}
	return ok
}

// Leq returns a bool array representing which elements of xs are <=
// than x0. It takes a output target as an optional argument to avoid excess
// allocations.
func Leq(xs []float64, x0 float64, out ...[]bool) []bool {
	ok := getOutput(out, len(xs))
	for i := range xs {
		ok[i] = xs[i] <= x0
	}
	return ok
}

// Geq returns a bool array representing which elements of xs are <=
// than x0. It takes a output target as an optional argument to avoid excess
// allocations.
func Geq(xs []float64, x0 float64, out ...[]bool) []bool {
	ok := getOutput(out, len(xs))
	for i := range xs {
		ok[i] = xs[i] >= x0
	}
	return ok
}

// IntGreater returns a bool array representing which elements of xs are greater
// than x0. It takes a output target as an optional argument to avoid excess
// allocations.
func IntGreater(xs []int, x0 int, out ...[]bool) []bool {
	ok := getOutput(out, len(xs))
	for i := range xs {
		ok[i] = xs[i] > x0
	}
	return ok
}

// IntLess returns a bool array representing which elements of xs are less
// than x0. It takes a output target as an optional argument to avoid excess
// allocations.
func IntLess(xs []int, x0 int, out ...[]bool) []bool {
	ok := getOutput(out, len(xs))
	for i := range xs {
		ok[i] = xs[i] < x0
	}
	return ok
}

// IntLeq returns a bool array representing which elements of xs are <=
// than x0. It takes a output target as an optional argument to avoid excess
// allocations.
func IntLeq(xs []int, x0 int, out ...[]bool) []bool {
	ok := getOutput(out, len(xs))
	for i := range xs {
		ok[i] = xs[i] <= x0
	}
	return ok
}

// IntGeq returns a bool array representing which elements of xs are <=
// than x0. It takes a output target as an optional argument to avoid excess
// allocations.
func IntGeq(xs []int, x0 int, out ...[]bool) []bool {
	ok := getOutput(out, len(xs))
	for i := range xs {
		ok[i] = xs[i] >= x0
	}
	return ok
}
