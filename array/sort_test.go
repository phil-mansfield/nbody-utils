package array

import (
	"math/rand"
	"sort"
	"testing"
)

func sliceEq(xs, ys []float64) bool {
	if len(xs) != len(ys) {
		return false
	}
	for i := range xs {
		if xs[i] != ys[i] {
			return false
		}
	}

	return true
}

func randSlice(n int) []float64 {
	xs := make([]float64, n)
	for i := range xs {
		xs[i] = rand.Float64()
	}
	return xs
}

func TestReverse(t *testing.T) {
	if !sliceEq([]float64{1, 2, 3, 4, 5}, Reverse([]float64{5, 4, 3, 2, 1})) ||
		!sliceEq([]float64{2, 3, 4, 5}, Reverse([]float64{5, 4, 3, 2})) {
		t.Errorf("Welp, I hope you're proud of yourself.")
	}
}

func BenchmarkReverse10(b *testing.B) {
	xs := make([]float64, 10)
	b.SetBytes(80)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Reverse(xs)
	}
}

func BenchmarkReverse1000(b *testing.B) {
	xs := make([]float64, 1000)
	b.SetBytes(8000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Reverse(xs)
	}
}

func BenchmarkReverse1000000(b *testing.B) {
	xs := make([]float64, 1000000)
	b.SetBytes(8000000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Reverse(xs)
	}
}

func BenchmarkShell10(b *testing.B) {
	xs := randSlice(10)
	buf := make([]float64, 10)

	b.SetBytes(80)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		ShellSort(buf)
	}
}

func BenchmarkShell100(b *testing.B) {
	xs := randSlice(100)
	b.SetBytes(800)
	b.ResetTimer()
	buf := make([]float64, 100)
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		ShellSort(buf)
	}
}

func BenchmarkShell1000(b *testing.B) {
	xs := randSlice(1000)
	b.SetBytes(8000)
	b.ResetTimer()
	buf := make([]float64, 1000)
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		ShellSort(buf)
	}
}

func BenchmarkShell10000(b *testing.B) {
	xs := randSlice(10000)
	buf := make([]float64, 10000)
	b.SetBytes(80000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		ShellSort(buf)
	}
}

func BenchmarkQuick10(b *testing.B) {
	xs := randSlice(10)
	b.SetBytes(80)
	b.ResetTimer()
	buf := make([]float64, 10)
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		QuickSort(buf)
	}
}

func BenchmarkQuick100(b *testing.B) {
	xs := randSlice(100)
	buf := make([]float64, 100)
	b.SetBytes(800)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		QuickSort(buf)
	}
}

func BenchmarkQuick1000(b *testing.B) {
	xs := randSlice(1000)
	buf := make([]float64, 1000)
	b.SetBytes(8000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		QuickSort(buf)
	}
}

func BenchmarkQuick10000(b *testing.B) {
	xs := randSlice(10000)
	buf := make([]float64, 10000)
	b.SetBytes(80000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		QuickSort(buf)
	}
}

func BenchmarkQuick100000(b *testing.B) {
	xs := randSlice(100000)
	buf := make([]float64, 100000)
	b.SetBytes(800000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		QuickSort(buf)
	}
}

func BenchmarkQuickIndex10(b *testing.B) {
	xs := randSlice(10)
	b.SetBytes(80)
	b.ResetTimer()
	buf := make([]float64, 10)
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		QuickSortIndex(buf)
	}
}

func BenchmarkQuickIndex100(b *testing.B) {
	xs := randSlice(100)
	buf := make([]float64, 100)
	b.SetBytes(800)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		QuickSortIndex(buf)
	}
}

func BenchmarkQuickIndex1000(b *testing.B) {
	xs := randSlice(1000)
	buf := make([]float64, 1000)
	b.SetBytes(8000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		QuickSortIndex(buf)
	}
}

func BenchmarkQuickIndex10000(b *testing.B) {
	xs := randSlice(10000)
	buf := make([]float64, 10000)
	b.SetBytes(80000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		QuickSortIndex(buf)
	}
}

func BenchmarkQuickIndex100000(b *testing.B) {
	xs := randSlice(100000)
	buf := make([]float64, 100000)
	b.SetBytes(800000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		QuickSortIndex(buf)
	}
}

func BenchmarkGo10(b *testing.B) {
	xs := randSlice(10)
	b.SetBytes(80)
	b.ResetTimer()
	buf := make([]float64, 10)
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		sort.Float64s(buf)
	}
}

func BenchmarkGo100(b *testing.B) {
	xs := randSlice(100)
	buf := make([]float64, 100)
	b.SetBytes(800)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		sort.Float64s(buf)
	}
}

func BenchmarkGo1000(b *testing.B) {
	xs := randSlice(1000)
	buf := make([]float64, 1000)
	b.SetBytes(8000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		sort.Float64s(buf)
	}
}

func BenchmarkGo10000(b *testing.B) {
	xs := randSlice(10000)
	buf := make([]float64, 10000)
	b.SetBytes(80000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, xs)
		sort.Float64s(buf)
	}
}

func BenchmarkMedian10(b *testing.B) {
	xs := randSlice(10)
	buf := make([]float64, len(xs))
	n := len(xs) / 2
	b.SetBytes(80)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NthLargest(xs, n, buf)
	}
}

func BenchmarkMedian100(b *testing.B) {
	xs := randSlice(100)
	buf := make([]float64, len(xs))
	n := len(xs) / 2
	b.SetBytes(800)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NthLargest(xs, n, buf)
	}
}

func BenchmarkMedian1000(b *testing.B) {
	xs := randSlice(1000)
	buf := make([]float64, len(xs))
	n := len(xs) / 2
	b.SetBytes(8000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NthLargest(xs, n, buf)
	}
}

func BenchmarkMedian10000(b *testing.B) {
	xs := randSlice(10000)
	buf := make([]float64, len(xs))
	n := len(xs) / 2
	b.SetBytes(80000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NthLargest(xs, n, buf)
	}
}

// Tests

func TestShellSort(t *testing.T) {
	for i := 0; i < 10; i++ {
		xs := randSlice(1000)
		ShellSort(xs)
		if !sort.Float64sAreSorted(xs) {
			t.Errorf("Failed to sort.")
		}
	}
}

func TestQuickSort(t *testing.T) {
	for i := 0; i < 10; i++ {
		xs := randSlice(1000)
		QuickSort(xs)
		if !sort.Float64sAreSorted(xs) {
			t.Errorf("Failed to sort.")
		}
	}
}

func TestQuickSortIndex(t *testing.T) {
	for i := 0; i < 10; i++ {
		xs := randSlice(1000)
		idx := QuickSortIndex(xs)
		sorted := make([]float64, len(xs))
		for i := range sorted {
			sorted[i] = xs[idx[i]]
		}
		
		if !sort.Float64sAreSorted(sorted) {
			t.Errorf("Failed to sort %.3g, %d, %.3g.", xs, idx, sorted)
		}
	}
}

func TestMedian(t *testing.T) {
	buf := make([]float64, 1000)
	for i := 0; i < 10; i++ {
		xs := randSlice(len(buf))
		QuickSort(xs)

		perm := rand.Perm(len(buf))
		mixed := make([]float64, len(buf))
		for j := range mixed {
			mixed[j] = xs[perm[j]]
		}

		for j := 1; j <= len(buf); j++ {
			val := NthLargest(mixed, j, buf)
			if val != xs[len(xs)-j] {
				t.Errorf("Failed to find NthLargest.")
			}
		}
	}
}
