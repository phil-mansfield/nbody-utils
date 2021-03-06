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

// IntReverse reverses a slice in place (and returns it for convenience).
func IntReverse(xs []int) []int {
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

// getFloatOutput is a utility function that gets the output array from an
// optional argument or allocates a new one.
func getFloatOutput(out [][]float64, n int) []float64 {
	if len(out) == 0 {
		return make([]float64, n)
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

// getIntOutput is a utility function that gets the output array from an
// optional argument or allocates a new one.
func getIntOutput(out [][]int, n int) []int {
	if len(out) == 0 {
		return make([]int, n)
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

// And returns a bool array corresponding an element-by-element && applied to
// all input arrays. Unlike other functions here it can't take an optional
// output argument because of the variadic input.
func And(xs ...[]bool) []bool {
	if len(xs) == 0 {
		panic("No input given to And.")
	}

	out := make([]bool, len(xs[0]))
	for i := range out {
		out[i] = true
	}

	for j := range xs {
		if len(xs[j]) != len(out) {
			panic(fmt.Sprintf("Argument %d of And() has length %d, not %d.",
				j, len(out), len(xs[j])))
		}
		for i := range out {
			out[i] = out[i] && xs[j][i]
		}
	}

	return out
}

// Or returns a bool array corresponding an element-by-element || applied to
// all input arrays. Unlike other functions here it can't take an optional
// output argument because of the variadic input.
func Or(xs ...[]bool) []bool {
	if len(xs) == 0 {
		panic("No input given to And.")
	}

	out := make([]bool, len(xs[0]))

	for j := range xs {
		if len(xs[j]) != len(out) {
			panic(fmt.Sprintf("Argument %d of And() has length %d, not %d.",
				j, len(out), len(xs[j])))
		}
		for i := range out {
			out[i] = out[i] || xs[j][i]
		}
	}

	return out
}

// Not applies element-by-element ! to an input array. It takes an optional
// output array.
func Not(xs []bool, out ...[]bool) []bool {
	ok := getOutput(out, len(xs))
	for i := range xs {
		ok[i] = !xs[i]
	}
	return ok
}

// Order reorders xs accoridng to the indiced in order (i.e. out[i] =
// xs[order[i]]). It takes an optional output argument. Behavior is undefined
// if out[0] == xs.
func Order(xs []float64, order []int, out ...[]float64) []float64 {
	ys := getFloatOutput(out, len(xs))
	
	if len(xs) != len(order) {
		panic(fmt.Sprintf("len(xs) = %d, but len(order) = %d",
			len(xs), len(order)))
	}

	for i := range order {
		ys[i] = xs[order[i]]
	}
	
	return ys
}

// IntOrder reorders xs accoridng to the indiced in order (i.e. out[i] =
// xs[order[i]]). It takes an optional output argument. Behavior is undefined
// if out[0] == xs.
func IntOrder(xs []int, order []int, out ...[]int) []int {
	ys := getIntOutput(out, len(xs))

	if len(xs) != len(order) {
		panic(fmt.Sprintf("len(xs) = %d, but len(order) = %d",
			len(xs), len(order)))
	}

	for i := range order {
		ys[i] = xs[order[i]]
	}

	return ys
}

// Cut applies a cut to xs such that order is preserved but all elements which
// are false in the ok array are removed. Does not take an output argument.
func Cut(xs []float64, ok []bool) []float64 {
	if len(xs) != len(ok) {
		panic(fmt.Sprintf("len(xs) = %d, but len(ok) = %d.",
			len(xs), len(ok)))
	}

	n := 0
	for i := range ok {
		if ok[i] { n++ }
	}

	out, j := make([]float64, n), 0
	for i := range ok {
		if ok[i] {
			out[j] = xs[i]
			j++
		}
	}

	return out
}

// IntCut applies a cut to xs such that order is preserved but all elements
// which are false in the ok array are removed. Does not take an output
// argument.
func IntCut(xs []int, ok []bool) []int {
	if len(xs) != len(ok) {
		panic(fmt.Sprintf("len(xs) = %d, but len(ok) = %d.",
			len(xs), len(ok)))
	}

	n := 0
	for i := range ok {
		if ok[i] { n++ }
	}

	out, j := make([]int, n), 0
	for i := range ok {
		if ok[i] {
			out[j] = xs[i]
			j++
		}
	}

	return out
}
