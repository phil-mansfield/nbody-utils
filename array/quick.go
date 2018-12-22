package array

const (
	manualLen = 25
)

// sort3 sorts three values from largest to smallest.
func sort3(x, y, z float64) (max, mid, min float64) {
	if x > y {
		if x > z {
			if y > z {
				return x, y, z
			} else {
				return x, z, y
			}
		} else {
			return z, x, y
		}
	} else {
		if y > z {
			if x > z {
				return y, x, z
			} else {
				return y, z, x
			}
		} else {
			return z, y, x
		}
	}
}

// QuickSort sorts an array in place via quicksort (and returns the result for
// convenience.)
//
// QuickSort is significantly faster than the standard library's quicksort on
// arrays of floats.
func QuickSort(xs []float64) []float64 {
	if len(xs) < manualLen {
		return ShellSort(xs)
	} else {
		pivIdx := partition(xs)
		QuickSort(xs[0:pivIdx])
		QuickSort(xs[pivIdx:len(xs)])
		return xs
	}
}

// partition rearranges the elements of a slice, xs, into two contiguous
// groups, such that every element of the first group is smaller than every
// element of the second. partition then returns the length of the first group.
func partition(xs []float64) int {
	n, n2 := len(xs), len(xs)/2
	// Take three values. The median will be the pivot, the other two will
	// be sentinel values so that we can avoid bounds checks.
	max, mid, min := sort3(xs[0], xs[n2], xs[n-1])
	xs[0], xs[n2], xs[n-1] = min, mid, max
	xs[1], xs[n2] = xs[n2], xs[1]

	lo, hi := 1, n-1
	for {
		lo++
		for xs[lo] < mid {
			lo++
		}
		hi--
		for xs[hi] > mid {
			hi--
		}
		if hi < lo {
			break
		}
		xs[lo], xs[hi] = xs[hi], xs[lo]
	}

	// Swap the pivot into the middle
	xs[1], xs[hi] = xs[hi], xs[1]

	return hi
}

// QuickSortIndex returns the indices of the array elements after they've been
// sorted in ascending order.
func QuickSortIndex(xs []float64) []int {
	xCopy := make([]float64, len(xs))
	for i := range xCopy { xCopy[i] = xs[i] }
	
	idx := make([]int, len(xs))
	for i := range idx { idx[i] = i }
	quickIndex(xCopy, idx)

	return idx
}

func sort3Index(x, y, z float64, ix, iy, iz int) (
	max, mid, min float64, maxi, midi, mini int,
) {
	if x > y {
		if x > z {
			if y > z {
				return x, y, z, ix, iy, iz
			} else {
				return x, z, y, ix, iz, iy
			}
		} else {
			return z, x, y, iz, ix, iy
		}
	} else {
		if y > z {
			if x > z {
				return y, x, z, iy, ix, iz
			} else {
				return y, z, x, iy, iz, ix
			}
		} else {
			return z, y, x, iz, iy, ix
		}
	}
}


func quickIndex(xs []float64, idx []int) {
	if len(idx) < manualLen {
		shellSortIndex(xs, idx)
	} else {
		pivIdx := partitionIndex(xs, idx)
		quickIndex(xs, idx[0:pivIdx])
		quickIndex(xs, idx[pivIdx:len(idx)])
	}	
}

func partitionIndex(xs []float64, idx []int) int {
	n, n2 := len(idx), len(idx)/2
	// Take three values. The median will be the pivot, the other two will
	// be sentinel values so that we can avoid bounds checks.
	_, _, _, maxi, midi, mini := sort3Index(
		xs[idx[0]], xs[idx[n2]], xs[idx[n-1]],
		idx[0], idx[n2], idx[n-1],
	)

	idx[0], idx[n2], idx[n-1] = mini, midi, maxi
	idx[1], idx[n2] = idx[n2], idx[1]
	lo, hi := 1, n-1
	for {
		lo++
		for xs[idx[lo]] < xs[midi] {
			lo++
		}
		hi--
		for xs[idx[hi]] > xs[midi] {
			hi--
		}
		if hi < lo {
			break
		}
		idx[lo], idx[hi] = idx[hi], idx[lo]
	}

	// Swap the pivot into the middle
	idx[1], idx[hi] = idx[hi], idx[1]

	return hi
}
