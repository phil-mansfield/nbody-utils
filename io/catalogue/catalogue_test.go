package catalogue

import (
	"testing"
)

func TestBlockSize(t *testing.T) {
	config := DefaultConfig
	config.MaxLineSize = 5
	config.MaxBlockSize = 15

	rd := TextFile("test_files/block_size_test.txt", config)
	txt, _ := rd.(*textReader)

	targetStarts := []int{0, 14, 29, 44}
	targetEnds := []int{14, 29, 44, 50}

	if !intsEq(txt.blockStarts, targetStarts) {
		t.Errorf("textReader.blockStarts = %d, but should be %d",
			txt.blockStarts, targetStarts)
	}

	if !intsEq(txt.blockEnds, targetEnds) {
		t.Errorf("textReader.blockEnds = %d, but should be %d",
			txt.blockEnds, targetEnds)
	}
}


func TestInts(t *testing.T) {
	rd := TextFile("test_files/int_test.txt")

	col0 := []int{10, 11, 12, 13, 14, 15, 16}
	col1 := []int{20, 21, 22, 23, 24, 25, 26}
	col2 := []int{30, 31, 32, 33, 34, 35, 36}
	col3 := []int{40, 41, 42, 43, 44, 45, 46}

	cols := rd.ReadIntBlock([]int{0, 3, 2, 0, 1}, 0)

	if !intsEq(col0, cols[0]) {
		t.Errorf("Read %d, but wanted %d", cols[0], col0)
	} else if !intsEq(col3, cols[1]) {
		t.Errorf("Read %d, but wanted %d", cols[1], col3)
	} else if !intsEq(col2, cols[2]) {
		t.Errorf("Read %d, but wanted %d", cols[2], col2)
	} else if !intsEq(col0, cols[3]) {
		t.Errorf("Read %d, but wanted %d", cols[3], col0)
	} else if !intsEq(col1, cols[4]) {
		t.Errorf("Read %d, but wanted %d", cols[4], col1)
	}

	config := DefaultConfig
	config.MaxLineSize = 2
	config.MaxBlockSize = 10

	rd = TextFile("test_files/int_block_test.txt", config)

	for i := 0; i < rd.Blocks(); i++ {
		col := rd.ReadIntBlock([]int{0}, i)[0]
		if len(col) != 5 { t.Errorf("Block %d has length %d", i, len(col)) }
	}

	col := rd.ReadInts([]int{0})[0]
	target := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	if !intsEq(col, target) {
		t.Errorf("Read %d, but wanted %d", col, target)
	}
}

func TestFloat64s(t *testing.T) {
	rd := TextFile("test_files/int_test.txt")

	col0 := []float64{10, 11, 12, 13, 14, 15, 16}
	col1 := []float64{20, 21, 22, 23, 24, 25, 26}
	col2 := []float64{30, 31, 32, 33, 34, 35, 36}
	col3 := []float64{40, 41, 42, 43, 44, 45, 46}

	cols := rd.ReadFloat64Block([]int{0, 3, 2, 0, 1}, 0)

	if !float64sEq(col0, cols[0]) {
		t.Errorf("Read %g, but wanted %g", cols[0], col0)
	} else if !float64sEq(col3, cols[1]) {
		t.Errorf("Read %g, but wanted %g", cols[1], col3)
	} else if !float64sEq(col2, cols[2]) {
		t.Errorf("Read %g, but wanted %g", cols[2], col2)
	} else if !float64sEq(col0, cols[3]) {
		t.Errorf("Read %g, but wanted %g", cols[3], col0)
	} else if !float64sEq(col1, cols[4]) {
		t.Errorf("Read %g, but wanted %g", cols[4], col1)
	}

	config := DefaultConfig
	config.MaxLineSize = 2
	config.MaxBlockSize = 10

	rd = TextFile("test_files/int_block_test.txt", config)

	for i := 0; i < rd.Blocks(); i++ {
		col := rd.ReadFloat64Block([]int{0}, i)[0]
		if len(col) != 5 { t.Errorf("Block %d has length %d", i, len(col)) }
	}

	col := rd.ReadFloat64s([]int{0})[0]
	target := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	if !float64sEq(col, target) {
		t.Errorf("Read %g, but wanted %g", col, target)
	}
}

func TestFloat32s(t *testing.T) {
	rd := TextFile("test_files/int_test.txt")

	col0 := []float32{10, 11, 12, 13, 14, 15, 16}
	col1 := []float32{20, 21, 22, 23, 24, 25, 26}
	col2 := []float32{30, 31, 32, 33, 34, 35, 36}
	col3 := []float32{40, 41, 42, 43, 44, 45, 46}

	cols := rd.ReadFloat32Block([]int{0, 3, 2, 0, 1}, 0)

	if !float32sEq(col0, cols[0]) {
		t.Errorf("Read %g, but wanted %g", cols[0], col0)
	} else if !float32sEq(col3, cols[1]) {
		t.Errorf("Read %g, but wanted %g", cols[1], col3)
	} else if !float32sEq(col2, cols[2]) {
		t.Errorf("Read %g, but wanted %g", cols[2], col2)
	} else if !float32sEq(col0, cols[3]) {
		t.Errorf("Read %g, but wanted %g", cols[3], col0)
	} else if !float32sEq(col1, cols[4]) {
		t.Errorf("Read %g, but wanted %g", cols[4], col1)
	}

	config := DefaultConfig
	config.MaxLineSize = 2
	config.MaxBlockSize = 10

	rd = TextFile("test_files/int_block_test.txt", config)

	for i := 0; i < rd.Blocks(); i++ {
		col := rd.ReadFloat32Block([]int{0}, i)[0]
		if len(col) != 5 { t.Errorf("Block %d has length %d", i, len(col)) }
	}

	col := rd.ReadFloat32s([]int{0})[0]
	target := []float32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	if !float32sEq(col, target) {
		t.Errorf("Read %g, but wanted %g", col, target)
	}
}

func intsEq(x, y []int) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}


func float64sEq(x, y []float64) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}


func float32sEq(x, y []float32) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}
