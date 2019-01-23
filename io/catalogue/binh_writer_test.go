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
