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

func intSliceEq(xs, ys []int) bool {
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

func TestIntReverse(t *testing.T) {
	if !sliceEq([]int{1, 2, 3, 4, 5}, IntReverse([]int{5, 4, 3, 2, 1})) ||
		!sliceEq([]int{2, 3, 4, 5}, IntReverse([]int{5, 4, 3, 2})) {
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

func TestAnd(t * testing.T) {
	x1 := []bool{true, true, false, false}
	x2 := []bool{false, true, true, true}
	x3 := []bool{true, true, true, false}

	res := []bool{false, true, false, false}
	out := And(x1, x2, x3)

	if !boolSliceEq(out, res) {
		t.Errorf("And(%v, %v, %v) = %v, not %v.", x1, x2, x3, out, res)
	}
}

func TestOr(t * testing.T) {
	x1 := []bool{true, true, false, false}
	x2 := []bool{false, true, true, false}
	x3 := []bool{true, true, true, false}

	res := []bool{true, true, true, false}
	out := Or(x1, x2, x3)

	if !boolSliceEq(out, res) {
		t.Errorf("And(%v, %v, %v) = %v, not %v.", x1, x2, x3, out, res)
	}
}

func TestNot(t *testing.T) {
	xs := []bool{true, false, true, false, true}
	res := []bool{false, true, false, true, false}
	
	ok := Not(xs)
	if !boolSliceEq(ok, res) {
		t.Errorf("Not(%v) = %v, not %v.", xs, ok, res)
	}

	out := make([]bool, 5)
	ok = Not(xs, out)
	if !boolSliceEq(out, res) || !boolSliceEq(out, ok) {
		t.Errorf("Not(%v) = %v, not %v.", xs, ok, res)
	}	
}

func TestOrder(t *testing.T) {
	xs := []float64{4, 5, 2, 0, 1, 3}
	order := []int{4, 5, 2, 0, 1, 3}
	res := []float64{0, 1, 2, 3, 4, 5}
	
	ys := Order(xs, order)
	if !sliceEq(ys, res) {
		t.Errorf("Order(%g, %d) = %g, not %g.", xs, order, ys, res)
	}

	out := make([]float64, 6)
	ys = Order(xs, order, out)
	if !sliceEq(out, res) || !sliceEq(out, ys) {
		t.Errorf("Order(%g, %d) = %g, not %g.", xs, order, ys, res)
	}	
}

func TestIntOrder(t *testing.T) {
	xs := []int{4, 5, 2, 0, 1, 3}
	order := []int{4, 5, 2, 0, 1, 3}
	res := []int{0, 1, 2, 3, 4, 5}
	
	ys := IntOrder(xs, order)
	if !intSliceEq(ys, res) {
		t.Errorf("Order(%g, %d) = %g, not %g.", xs, order, ys, res)
	}

	out := make([]int, 6)
	ys = IntOrder(xs, order, out)
	if !intSliceEq(out, res) || !intSliceEq(out, ys) {
		t.Errorf("Order(%g, %d) = %g, not %g.", xs, order, ys, res)
	}	
}


func TestCut(t *testing.T) {
	xs := []float64{1, 2, 3, 4, 5}
	ok := []bool{true, true, false, false, true}
	res := []float64{1, 2, 5}
	out := Cut(xs, ok)

	if !sliceEq(out, res) {
		t.Errorf("Cut(%g, %v) = %g, not %g", xs, ok, out, res)
	}
}

func TestIntCut(t *testing.T) {
	xs := []int{1, 2, 3, 4, 5}
	ok := []bool{true, true, false, false, true}
	res := []int{1, 2, 5}
	out := IntCut(xs, ok)

	if !intSliceEq(out, res) {
		t.Errorf("Cut(%d, %v) = %g, not %g", xs, ok, out, res)
	}
}
