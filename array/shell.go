package array

var (
	shellk = 0
	indexk = 0
)

// ShellSort sorts an array in place via Shell's method (and returns the
// result for convenience).
func ShellSort(xs []float64) []float64 {
	n := len(xs)
	if n == 1 {
		return xs
	}

	inc := 1
	for inc <= n {
		inc = inc*3 + 1
	}
	
	for inc > 1 {
		inc /= 3
		for i := inc; i < n; i++ {
			v := xs[i]
			j := i
			for xs[j-inc] > v {
				xs[j] = xs[j-inc]
				j -= inc
				if j < inc {
					break
				}
			}
			xs[j] = v
		}
	}
	return xs
}

func ShellSortIndex(xs []float64) []int {
	idx := make([]int, len(xs))
	for i := range idx { idx[i] = i }
	shellSortIndex(xs, idx)
	return idx
}

// shellSortIndex does an in-place Shell sort of idx such that xs[idx[i]] is
// sorted in ascending order.
func shellSortIndex(xs []float64, idx []int) {
	n := len(idx)
	if n == 1 { return }

	inc := 1
	for inc <= n { inc = inc*3 + 1 }

	for inc > 1 {
		inc /= 3
		for i := inc; i < n; i++ {
			v := xs[idx[i]]
			vi := idx[i]
			
			j := i
			for xs[idx[j-inc]] > v {
				idx[j] = idx[j - inc]
				j -= inc
				if j < inc { break }
			}
			idx[j] = vi
		}
	}
	return
}
