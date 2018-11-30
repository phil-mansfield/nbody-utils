package array

import (
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

func boolSliceEq(xs, ys []bool) bool {
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

func TestGreater(t *testing.T) {
	xs, x0 := []float64{1, 2, 4, 3, 5}, 3.0
	res := []bool{false, false, true, false, true}
	
	ok := Greater(xs, x0)
	if !boolSliceEq(ok, res) {
		t.Errorf("Greater(%g, %g) = %v, not %v.", xs, x0, ok, res)
	}

	out := make([]bool, 5)
	ok = Greater(xs, x0, out)
	if !boolSliceEq(out, res) || !boolSliceEq(out, ok) {
		t.Errorf("Greater(%g, %g) = %v, not %v.", xs, x0, ok, res)
	}	
}

func TestLess(t *testing.T) {
	xs, x0 := []float64{1, 2, 4, 3, 5}, 3.0
	res := []bool{true, true, false, false, false}
	
	ok := Less(xs, x0)
	if !boolSliceEq(ok, res) {
		t.Errorf("Less(%g, %g) = %v, not %v.", xs, x0, ok, res)
	}

	out := make([]bool, 5)
	ok = Less(xs, x0, out)
	if !boolSliceEq(out, res) || !boolSliceEq(out, ok) {
		t.Errorf("Less(%g, %g) = %v, not %v.", xs, x0, ok, res)
	}	
}

func TestGeq(t *testing.T) {
	xs, x0 := []float64{1, 2, 4, 3, 5}, 3.0
	res := []bool{false, false, true, true, true}
	
	ok := Geq(xs, x0)
	if !boolSliceEq(ok, res) {
		t.Errorf("Geq(%g, %g) = %v, not %v.", xs, x0, ok, res)
	}

	out := make([]bool, 5)
	ok = Geq(xs, x0, out)
	if !boolSliceEq(out, res) || !boolSliceEq(out, ok) {
		t.Errorf("Geq(%g, %g) = %v, not %v.", xs, x0, ok, res)
	}	
}

func TestLeq(t *testing.T) {
	xs, x0 := []float64{1, 2, 4, 3, 5}, 3.0
	res := []bool{true, true, false, true, false}
	
	ok := Leq(xs, x0)
	if !boolSliceEq(ok, res) {
		t.Errorf("Leq(%g, %g) = %v, not %v.", xs, x0, ok, res)
	}

	out := make([]bool, 5)
	ok = Leq(xs, x0, out)
	if !boolSliceEq(out, res) || !boolSliceEq(out, ok) {
		t.Errorf("Leq(%g, %g) = %v, not %v.", xs, x0, ok, res)
	}	
}

func TestIntGreater(t *testing.T) {
	xs, x0 := []int{1, 2, 4, 3, 5}, 3
	res := []bool{false, false, true, false, true}
	
	ok := IntGreater(xs, x0)
	if !boolSliceEq(ok, res) {
		t.Errorf("IntGreater(%g, %g) = %v, not %v.", xs, x0, ok, res)
	}

	out := make([]bool, 5)
	ok = IntGreater(xs, x0, out)
	if !boolSliceEq(out, res) || !boolSliceEq(out, ok) {
		t.Errorf("IntGreater(%g, %g) = %v, not %v.", xs, x0, ok, res)
	}	
}

func TestIntLess(t *testing.T) {
	xs, x0 := []int{1, 2, 4, 3, 5}, 3
	res := []bool{true, true, false, false, false}
	
	ok := IntLess(xs, x0)
	if !boolSliceEq(ok, res) {
		t.Errorf("IntLess(%g, %g) = %v, not %v.", xs, x0, ok, res)
	}

	out := make([]bool, 5)
	ok = IntLess(xs, x0, out)
	if !boolSliceEq(out, res) || !boolSliceEq(out, ok) {
		t.Errorf("IntLess(%g, %g) = %v, not %v.", xs, x0, ok, res)
	}	
}

func TestIntGeq(t *testing.T) {
	xs, x0 := []int{1, 2, 4, 3, 5}, 3
	res := []bool{false, false, true, true, true}
	
	ok := IntGeq(xs, x0)
	if !boolSliceEq(ok, res) {
		t.Errorf("IntGeq(%g, %g) = %v, not %v.", xs, x0, ok, res)
	}

	out := make([]bool, 5)
	ok = IntGeq(xs, x0, out)
	if !boolSliceEq(out, res) || !boolSliceEq(out, ok) {
		t.Errorf("IntGeq(%g, %g) = %v, not %v.", xs, x0, ok, res)
	}	
}

func TestIntLeq(t *testing.T) {
	xs, x0 := []int{1, 2, 4, 3, 5}, 3
	res := []bool{true, true, false, true, false}
	
	ok := IntLeq(xs, x0)
	if !boolSliceEq(ok, res) {
		t.Errorf("IntLeq(%g, %g) = %v, not %v.", xs, x0, ok, res)
	}

	out := make([]bool, 5)
	ok = IntLeq(xs, x0, out)
	if !boolSliceEq(out, res) || !boolSliceEq(out, ok) {
		t.Errorf("IntLeq(%g, %g) = %v, not %v.", xs, x0, ok, res)
	}	
}
