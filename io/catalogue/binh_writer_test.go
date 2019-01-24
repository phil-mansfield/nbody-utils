package catalogue

import (
	"math"
	"bytes"
	"testing"
)

func TestEncoderBuffers(t *testing.T) {
	enc := &BinhEncoder{ }

	ns := []int{10, 5, 1<<20, 0}
	for i := range ns {
		n := ns[i]

		buf8 := enc.int8Buffer(n)
		buf16 := enc.int16Buffer(n)
		buf32 := enc.int32Buffer(n)
		bufF32 := enc.float32Buffer(n)
		bufF64 := enc.float64Buffer(n)
		buf64 := enc.int64Buffer(n)

		if len(buf8) != n {
			t.Errorf("Expected int8 buffer of length %d, got %d", n, len(buf8))
		}
		if len(buf16) != n {
			t.Errorf("Expected int16 buffer of length %d, got %d", n,len(buf16))
		}
		if len(buf32) != n {
			t.Errorf("Expected int32 buffer of length %d, got %d", n,len(buf32))
		}
		if len(buf64) != n {
			t.Errorf("Expected int64 buffer of length %d, got %d", n,len(buf64))
		}
		if len(bufF32) != n {
			t.Errorf("Expected float32 buffer of length %d, got %d",
				n, len(bufF32))
		}
		if len(bufF64) != n {
			t.Errorf("Expected float64 buffer of length %d, got %d",
				n, len(bufF64))
		}
	}
}

func TestIntEncoder(t *testing.T) {
	data := []int{-10, 0, -20, 200}

	enc := &BinhEncoder{ }
	
	flags := []ColumnFlag{ Int64, Int32, Int16, Int8 }
	sizes := []int{ 32, 16, 8, 4} 

	for i := range flags {
		writeBuf := &bytes.Buffer{ }
		key := enc.EncodeInts(flags[i], data, writeBuf)

		if writeBuf.Len() != sizes[i] {
			t.Errorf("%d) Expected size = %d, but got %d.",
				i, sizes[i], writeBuf.Len())
		}

		readBuf := bytes.NewBuffer(writeBuf.Bytes())
		out := make([]int, len(data))
		enc.DecodeInts(flags[i], key, readBuf, out)

		if !intsEq(out, data) {
			t.Errorf("%d), Expected out = %d for flag %d, got %d.", i,
				data, flags[i], out)
		}
	}
}

func TestFloat64Encoder(t *testing.T) {
	data := []float64{1, 2, 4, 8}

	enc := &BinhEncoder{ }
	
	flags := []ColumnFlag{ Float64, Float32, QFloat64, QFloat32, QFloat16,
		QFloat8, QLogFloat64, QLogFloat32, QLogFloat16, QLogFloat8}
	sizes := []int{ 32, 16, 32, 16, 8, 4, 32, 16, 8, 4 }
	delta := 0.25

	for i := range flags {
		writeBuf := &bytes.Buffer{ }
		key := enc.EncodeFloat64s(flags[i], delta, data, writeBuf)

		if writeBuf.Len() != sizes[i] {
			t.Errorf("%d) Expected size = %d, but got %d.",
				i, sizes[i], writeBuf.Len())
		}

		readBuf := bytes.NewBuffer(writeBuf.Bytes())
		out := make([]float64, len(data))
		enc.DecodeFloat64s(flags[i], delta, key, readBuf, out)

		switch flags[i] {
		case QLogFloat64, QLogFloat32, QLogFloat16, QLogFloat8:
			lout := make([]float64, len(data))
			ldata := make([]float64, len(data))
			for i := range lout {
				lout[i], ldata[i] = math.Log10(out[i]), math.Log10(data[i])
			}
			if !float64sAlmostEq(lout, ldata, delta) {
				t.Errorf("%d), Expected log10(out) = %g for flag %d, got %g.",
					i, data, flags[i], out)
			}
		default:
			if !float64sAlmostEq(out, data, delta) {
				t.Errorf("%d), Expected out = %g for flag %d, got %g.", i,
					data, flags[i], out)
			}
		}
	}
}

func TestFloat32Encoder(t *testing.T) {
	data := []float32{1, 2, 4, 8}

	enc := &BinhEncoder{ }
	
	flags := []ColumnFlag{ Float64, Float32, QFloat64, QFloat32, QFloat16,
		QFloat8, QLogFloat64, QLogFloat32, QLogFloat16, QLogFloat8}
	sizes := []int{ 32, 16, 32, 16, 8, 4, 32, 16, 8, 4 }
	delta := float32(0.25)

	for i := range flags {
		writeBuf := &bytes.Buffer{ }
		key := enc.EncodeFloat32s(flags[i], delta, data, writeBuf)

		if writeBuf.Len() != sizes[i] {
			t.Errorf("%d) Expected size = %d, but got %d.",
				i, sizes[i], writeBuf.Len())
		}

		readBuf := bytes.NewBuffer(writeBuf.Bytes())
		out := make([]float32, len(data))
		enc.DecodeFloat32s(flags[i], delta, key, readBuf, out)

		switch flags[i] {
		case QLogFloat64, QLogFloat32, QLogFloat16, QLogFloat8:
			lout := make([]float64, len(data))
			ldata := make([]float64, len(data))
			for i := range lout {
				lout[i] = math.Log10(float64(out[i]))
				ldata[i] = math.Log10(float64(data[i]))
			}
			if !float64sAlmostEq(lout, ldata, float64(delta)) {
				t.Errorf("%d), Expected log10(out) = %g for flag %d, got %g.",
					i, data, flags[i], out)
			}
		default:
			if !float32sAlmostEq(out, data, delta) {
				t.Errorf("%d), Expected out = %g for flag %d, got %g.", i,
					data, flags[i], out)
			}
		}
	}
}

func TestIntColumnType(t *testing.T) {
	tests := []struct{
		x []int
		key int64
		flag ColumnFlag
	} {
		{[]int{0, 1}, 0, Int8},
		{[]int{10, 11}, 10, Int8},
		{[]int{-10, -9}, -10, Int8},

		{[]int{math.MinInt8, math.MaxInt8}, math.MinInt8, Int8},
		{[]int{0, math.MaxInt8 - math.MinInt8}, 0, Int8},
		{[]int{math.MinInt8, math.MaxInt8 + 1}, math.MinInt8, Int16},
		{[]int{0, math.MaxInt8 - math.MinInt8 + 1}, 0, Int16},

		{[]int{math.MinInt16, math.MaxInt16}, math.MinInt16, Int16},
		{[]int{0, math.MaxInt16 - math.MinInt16}, 0, Int16},
		{[]int{math.MinInt16, math.MaxInt16 + 1}, math.MinInt16, Int32},
		{[]int{0, math.MaxInt16 - math.MinInt16 + 1}, 0, Int32},

		{[]int{math.MinInt32, math.MaxInt32}, math.MinInt32, Int32},
		{[]int{0, math.MaxInt32 - math.MinInt32}, 0, Int32},
		{[]int{math.MinInt32, math.MaxInt32 + 1}, math.MinInt32, Int64},
		{[]int{0, math.MaxInt32 - math.MinInt32 + 1}, 0, Int64},

		{[]int{math.MinInt64, math.MaxInt64}, math.MinInt64, Int64},
	}

	for i := range tests {
		flag, key := intColumnType(tests[i].x)
 		if key != tests[i].key {
			t.Errorf("%d) Expected key %d for %d, got %d.",
				i, tests[i].key, tests[i].x, key)
		}
		if flag != tests[i].flag {
			t.Errorf("%d) Expected flag %d for %d, got %d.",
				i, tests[i].flag, tests[i].x, flag)
		}
	}
}

func TestFloat64ColumnType(t *testing.T) {
	tests := []struct{
		x []float64
		delta float64
		key int64
		flag ColumnFlag
	} {
		{[]float64{1.0, 2.0}, 0.0, 0, Float32},
		{[]float64{1.0, 2.0}, 0.1, 10, QFloat8},
		{[]float64{1.0, 2.0}, 1e-3, 1000, QFloat16},
		{[]float64{1.0, 2.0}, 1e-6, 0, Float32},
		{[]float64{1.0, 2.0}, 1e-10, 0, Float32},
	}

	for i := range tests {
		flag, key := float64ColumnType(tests[i].x, tests[i].delta)
 		if key != tests[i].key {
			t.Errorf("%d) Expected key %d for %g, delta=%g, got %g.",
				i, tests[i].key, tests[i].x, tests[i].delta, key)
		}
		if flag != tests[i].flag {
			t.Errorf("%d) Expected flag %d for %g, delta=%g, got %g.",
				i, tests[i].flag, tests[i].x, tests[i].delta, flag)
		}
	}
}

func TestLogFloat64ColumnType(t *testing.T) {
	tests := []struct{
		x []float64
		delta float64
		key int64
		flag ColumnFlag
	} {
		{[]float64{10.0, 100.0}, 0.0, 0, Float32},
		{[]float64{10.0, 100.0, 0.0}, 0.1, 0, Float32},
		{[]float64{10.0, 100.0}, 0.1, 10, QLogFloat8},
		{[]float64{10.0, 100.0}, 1e-3, 1000, QLogFloat16},
		{[]float64{10.0, 100.0}, 1e-6, 0, Float32},
		{[]float64{10.0, 100.0}, 1e-10, 0, Float32},
	}

	for i := range tests {
		flag, key := logFloat64ColumnType(tests[i].x, tests[i].delta)
 		if key != tests[i].key {
			t.Errorf("%d) Expected key %d for %g, delta=%g, got %d.",
				i, tests[i].key, tests[i].x, tests[i].delta, key)
		}
		if flag != tests[i].flag {
			t.Errorf("%d) Expected flag %d for %g, delta=%g, got %d.",
				i, tests[i].flag, tests[i].x, tests[i].delta, flag)
		}
	}
}


func TestBufferIndex(t *testing.T) {
	tests := []struct {
		isInt []bool
		bufIdx, icols, fcols []int
	} {
		{[]bool{true}, []int{0}, []int{0}, []int{}},
		{[]bool{false}, []int{0}, []int{}, []int{0}},
		{[]bool{true, true, true}, []int{0, 1, 2}, []int{0, 1, 2}, []int{}},
		{[]bool{false, false, false}, []int{0, 1, 2}, []int{}, []int{0, 1, 2}},
		{[]bool{false, false, true, false},
			[]int{0, 1, 0, 2}, []int{2}, []int{0, 1, 3}},
		{[]bool{true, true, false, true},
			[]int{0, 1, 0, 2}, []int{0, 1, 3}, []int{2}},
	}

	for i := range tests {
		bufIdx, icols, fcols := bufferIndex(tests[i].isInt)

		if !intsEq(tests[i].bufIdx, bufIdx) {
			t.Errorf("%d) expected bufIdx = %d, got %d",
				i, tests[i].bufIdx, bufIdx)
		}
		if !intsEq(tests[i].icols, icols) {
			t.Errorf("%d) expected icols = %d, got %d",
				i, tests[i].icols, icols)
		}
		if !intsEq(tests[i].fcols, fcols) {
			t.Errorf("%d) expected focls = %d, got %d",
				i, tests[i].fcols, fcols)
		}
	}
}

func float64sAlmostEq(x, y []float64, delta float64) bool {
	if len(x) != len(y) { return false }
	
	for i := range x {
		if x[i] - y[i] > delta || y[i] - x[i] > delta { return false }
	}

	return true
}

func float32sAlmostEq(x, y []float32, delta float32) bool {
	if len(x) != len(y) { return false }
	
	for i := range x {
		if x[i] - y[i] > delta || y[i] - x[i] > delta { return false }
	}

	return true
}
