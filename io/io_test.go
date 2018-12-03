package io

import (
	"testing"
)

func TestExpand(t *testing.T) {
	tests := []struct{ len, cap, newLen int } {
		{0, 0, 10},
		{0, 5, 10},
		{0, 5, 4},
		{0, 5, 5},
		{5, 5, 5},
		{4, 5, 4},
		{3, 5, 4},
		{4, 5, 3},
	}

	for i, test := range tests {
		xVec := make([][3]float64, test.len, test.cap)
		xFloat := make([]float64, test.len, test.cap)
		xInt := make([]int, test.len, test.cap)

		xVec = expandVec(xVec, test.newLen)
		xFloat = expandFloat(xFloat, test.newLen)
		xInt = expandInt(xInt, test.newLen)

		if len(xVec) != test.newLen {
			t.Errorf("Failed test %d. len(xVec) = %d.", i, len(xVec))
		}
		if len(xFloat) != test.newLen {
			t.Errorf("Failed test %d. len(xVec) = %d.", i, len(xFloat))
		}
		if len(xInt) != test.newLen {
			t.Errorf("Failed test %d. len(xVec) = %d.", i, len(xInt))
		}
	}
}
