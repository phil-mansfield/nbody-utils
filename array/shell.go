package array

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

// shellSortIndex does an in-place Shell sort of idx such that xs[idx[i]] is
// sorted in ascending order.
func shellSortIndex(xs []float64, idx []int) {
	n := len(xs)
	if n == 1 {
		return
	}

	inc := 1
	for inc <= n {
		inc = inc*3 + 1
	}

	for inc > 1 {
		inc /= 3
		for i := inc; i < n; i++ {
			v, vi := xs[i], idx[i]
			j := i
			for xs[j-inc] > v {
				xs[j] = xs[j-inc]
				idx[j] = idx[j-inc]
				j -= inc
				if j < inc {
					break
				}
			}
			xs[j] = v
			idx[j] = vi
		}
	}
}
