package catalogue

import (
	"bytes"
	"math"
	"os"
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

func TestParseColumnInfo(t *testing.T) {
	info := []string{
		"A : int ",
		"B : float ",
		"C : log ",
		"D : float: 2.0 ",
		"E : log : 2.0 ",
	}
	names := []string{"a", "b", "c", "d", "e"}
	isInt := []bool{true, false, false, false, false}
	isLog := []bool{false, false, true, false, true}
	deltas := []float64{0, 0, 0, 2, 2}

	outNames, outIsInt, outIsLog, outDeltas := parseColumnInfo(info)

	for i := range info {
		switch {
		case names[i] != outNames[i]:
			t.Errorf("%d) wanted name %s, got %s", i, names[i], outNames[i])
		case isInt[i] != outIsInt[i]:
			t.Errorf("%d) wanted name %s, got %s", i, isInt[i], outIsInt[i])
		case isLog[i] != outIsLog[i]:
			t.Errorf("%d) wanted name %s, got %s", i, isLog[i], outIsLog[i])
		case deltas[i] != outDeltas[i]:
			t.Errorf("%d) wanted name %s, got %s", i, deltas[i], outDeltas[i])
		}
	}
}

func TestNewBinhHeader(t *testing.T) {
	config := ParseBinhConfig("test_files/binh_test.config")
	hd := newBinhHeader("test_files/binh_test.txt", 2, config)

	textHeader := `# Header line 1....
# Header line 2....
# Header line 3....
`
	
	switch {
	case hd.Version != 2:
		t.Errorf("Version = %d, not %d", hd.Version, 2)
	case hd.Columns != 4:
		t.Errorf("Columns = %d, not %d", hd.Columns, 4)
	case hd.MassColumn != 1:
		t.Errorf("MassColumn = %d, not %d", hd.MassColumn, 1)
	case hd.Blocks != 2:
		t.Errorf("Blocks = %d, not %d", hd.Blocks, 2)
	case hd.TextHeaderLength != 60:
		t.Errorf("TextHeaderLength = %d, not %d", hd.TextHeaderLength, 60)
	case hd.TextColumnNamesLength != 15:
		t.Errorf("TextColumnNamesLength = %d, not %d", 
			hd.TextColumnNamesLength, 15)
	case hd.MinMass != 1e10:
		t.Errorf("MinMass = %g not %g", hd.MinMass, 1e10)
	case !float64sEq(hd.Deltas, []float64{0, 0.01, 0.01, 1}):
		t.Errorf("Deltas = %g, not %g", hd.Deltas,
			[]float64{0, 0.01, 0.01, 1})
	case string(hd.TextHeader) != textHeader:
		t.Errorf("Text header = '%s', not '%s'",
			string(hd.TextHeader), textHeader)
	case string(hd.TextColumnNames) != "id,mvir,m200m,x":
		t.Errorf("TextColumnNames = %s, not %s.",
			hd.TextColumnNames, "id,mvir,m200m,x")
	case !uint8Eq(hd.ColumnSkipped, []uint8{0, 0, 1, 0}):
		t.Errorf("ColumnSkipped = %v, not %v", hd.ColumnSkipped,
			[]uint8{0, 0, 1, 0})
	}
}

func TestReadBinhHeader(t *testing.T) {
	textConfig := DefaultConfig
	textConfig.MaxBlockSize = 100
	textConfig.MaxLineSize = 20

	config := ParseBinhConfig("test_files/binh_test.config")

	TextToBinh(
		"test_files/binh_test.txt",
		"test_files/binh_test.binh",
		config, textConfig,
	)

	rd, err := os.Open("test_files/binh_test.binh")
    if err != nil { panic(err.Error()) }

	hd := readBinhHeader(rd)

	textHeader := `# Header line 1....
# Header line 2....
# Header line 3....
`

	switch {
	case hd.Version != 2:
		t.Errorf("Version = %d, not %d", hd.Version, 2)
	case hd.Columns != 4:
		t.Errorf("Columns = %d, not %d", hd.Columns, 4)
	case hd.MassColumn != 1:
		t.Errorf("MassColumn = %d, not %d", hd.MassColumn, 1)
	case hd.Blocks != 2:
		t.Errorf("Blocks = %d, not %d", hd.Blocks, 2)
	case hd.TextHeaderLength != 60:
		t.Errorf("TextHeaderLength = %d, not %d", hd.TextHeaderLength, 60)
	case hd.TextColumnNamesLength != 15:
		t.Errorf("TextColumnNamesLength = %d, not %d", 
			hd.TextColumnNamesLength, 15)
	case hd.MinMass != 1e10:
		t.Errorf("MinMass = %g not %g", hd.MinMass, 1e10)
	case !float64sEq(hd.Deltas, []float64{0, 0.01, 0.01, 1}):
		t.Errorf("Deltas = %g, not %g", hd.Deltas,
			[]float64{0, 0.01, 0.01, 1})
	case string(hd.TextHeader) != textHeader:
		t.Errorf("Text header = '%s', not '%s'",
			string(hd.TextHeader), textHeader)
	case string(hd.TextColumnNames) != "id,mvir,m200m,x":
		t.Errorf("TextColumnNames = %s, not %s.",
			hd.TextColumnNames, "id,mvir,m200m,x")
	case !uint8Eq(hd.ColumnSkipped, []uint8{0, 0, 1, 0}):
		t.Errorf("ColumnSkipped = %v, not %v", hd.ColumnSkipped,
			[]uint8{0, 0, 1, 0})
	}
}

func TestBlockHeaderOffsets(t *testing.T) {
	textConfig := DefaultConfig
	textConfig.MaxBlockSize = 100
	textConfig.MaxLineSize = 20

	config := ParseBinhConfig("test_files/binh_test.config")

	TextToBinh(
		"test_files/binh_test.txt",
		"test_files/binh_test.binh",
		config, textConfig,
	)

	rd, err := os.Open("test_files/binh_test.binh")
    if err != nil { panic(err.Error()) }

	hd := readBinhHeader(rd)

	offsets := blockHeaderOffsets(rd, hd)
	if !intsEq(offsets, []int{175, 250}) {
		t.Errorf("Got block offsets = %d, expected %d",
			offsets, []int{175, 250})
	}
}

func TestNewBinhReader(t *testing.T) {
	textConfig := DefaultConfig
	textConfig.MaxBlockSize = 100
	textConfig.MaxLineSize = 20

	config := ParseBinhConfig("test_files/binh_test.config")

	TextToBinh(
		"test_files/binh_test.txt",
		"test_files/binh_test.binh",
		config, textConfig,
	)

	rd := newBinhReader("test_files/binh_test.binh")

	if !intsEq(rd.blockOffsets, []int{247, 322}) {
		t.Errorf("got blockOffsets = %d, not %d",
			rd.blockOffsets, []int{247, 322})
	}

	if rd.Blocks() != 2 {
		t.Errorf("got blocks = %d, not %d", rd.Blocks(), 2)
	}

	if rd.hd.Haloes != 4 {
		t.Errorf("got Haloes = %d, not %d", rd.hd.Haloes, 4)
	}

	if !int64sEq(rd.hd.BlockHaloes, []int64{1, 3}) {
		t.Errorf("got BlockHaloes = %d, not %d",
			rd.hd.BlockHaloes, []int{1, 3})
	}
}

func TestBinhReadFloat64s(t *testing.T) {
	textConfig := DefaultConfig
	textConfig.MaxBlockSize = 100
	textConfig.MaxLineSize = 20

	config := ParseBinhConfig("test_files/binh_test.config")

	TextToBinh(
		"test_files/binh_test.txt",
		"test_files/binh_test.binh",
		config, textConfig,
	)

	rd := newBinhReader("test_files/binh_test.binh")

	block0 := rd.ReadFloat64Block([]string{"mvir", "x"}, 0)
	block1 := rd.ReadFloat64Block([]string{"x", "mvir"}, 1)
	all := rd.ReadFloat64s([]string{"mvir", "x"})

	mvir0, x0 := block0[0], block0[1]
	x1, mvir1 := block1[0], block1[1]
	mvir, x := all[0], all[1]

	if !float64sAlmostEq(x, []float64{150, 125, 130, 100}, 1) {
		t.Errorf("Got %.3g, expected %.3g", x, []float64{150, 125, 130, 100})
	}
	if !float64sAlmostEq(x0, []float64{150}, 1) {
		t.Errorf("Got %.3g, expected %.3g", x0, []float64{150})
	}
	if !float64sAlmostEq(x1, []float64{125, 130, 100}, 1) {
		t.Errorf("Got %.3g, expected %.3g", x1, []float64{125, 130, 100})
	}
	if !logFloat64sAlmostEq(mvir, []float64{1e12, 1e10, 1e11, 1e13}, 0.01) {
		t.Errorf("Got %.3g, expected %.3g", mvir,
			[]float64{1e12, 1e10, 1e11, 1e13})
	}
	if !logFloat64sAlmostEq(mvir0, []float64{1e12}, 0.01) {
		t.Errorf("Got %.3g, expected %.3g", mvir0, []float64{1e12})
	}
	if !logFloat64sAlmostEq(mvir1, []float64{1e10, 1e11, 1e13}, 0.01) {
		t.Errorf("Got %.3g, expected %.3g", mvir1, []float64{1e13, 1e11, 1e10})
	}
}

func boolsEq(x, y []bool) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

func uint8Eq(x, y []uint8) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}

func int64sEq(x, y []int64) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}


func float64sAlmostEq(x, y []float64, delta float64) bool {
	if len(x) != len(y) { return false }
	
	for i := range x {
		if x[i] - y[i] > delta || y[i] - x[i] > delta { return false }
	}

	return true
}

func logFloat64sAlmostEq(x, y []float64, delta float64) bool {
	if len(x) != len(y) { return false }
	
	for i := range x {
		if math.Log10(x[i]) - math.Log10(y[i]) > delta ||
			math.Log10(y[i]) - math.Log10(x[i]) > delta {
			return false
		}
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
