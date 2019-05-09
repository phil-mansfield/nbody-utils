package snapshot

import (
	"bytes"
	"io/ioutil"
	"math"
	"os"
	"testing"

	"github.com/phil-mansfield/nbody-utils/container"
	
	"unsafe"
)

var testHeader = &lvecHeader{
	Magic: LVecMagicNumber,
	Version: LVecVersion,
	VarType: lvecX,
	Method: lvecBoxMethod,
	Idx: 1,
	Cells: 3,
	RawHeaderBytes: 5,
	SubCells: 2,
	SubCellVectorsMin: 0,
	SubCellVectorsBits: 0,
	BitsMin: 0,
	BitsBits: 0,
	Pix: 100,
	Limits: [2]float64{0, 50},
	Delta: 1.0,
	Offsets: [5]uint64{8 + uint64(unsafe.Sizeof(lvecHeader{})), 0, 0, 0},
	Hd: Header{ 1, 2, 3, 4, 5, 6, 7, 8, 9, 10 },
}

func lvecHeaderEq(hd1, hd2 *lvecHeader) bool {
	return hd1.Magic == hd2.Magic &&
		hd1.Version == hd2.Version &&
		hd1.VarType == hd2.VarType &&
		hd1.Method == hd2.Method &&
		hd1.Idx == hd2.Idx &&
		hd1.Cells == hd2.Cells &&
		hd1.SubCells == hd2.SubCells &&
		hd1.RawHeaderBytes == hd2.RawHeaderBytes &&
		hd1.SubCellVectorsMin == hd2.SubCellVectorsMin &&
		hd1.SubCellVectorsBits == hd2.SubCellVectorsBits &&
		hd1.BitsMin == hd2.BitsMin &&
		hd1.BitsBits == hd2.BitsBits &&
		hd1.Pix == hd2.Pix &&
		hd1.Limits == hd2.Limits &&
		hd1.Delta == hd2.Delta &&
		hd1.Offsets[0] == hd2.Offsets[0] &&
		hd1.Offsets[1] == hd2.Offsets[1] &&
		hd1.Offsets[2] == hd2.Offsets[2] &&
		hd1.Offsets[3] == hd2.Offsets[3] &&
		hd1.Offsets[4] == hd2.Offsets[4] &&
		headerEq(&hd1.Hd, &hd2.Hd)
}

func headerEq(hd1, hd2 *Header) bool {
	return hd1.Z == hd2.Z &&
		hd1.Scale == hd2.Scale &&
		hd1.OmegaM == hd2.OmegaM &&
		hd1.OmegaL == hd2.OmegaL &&
		hd1.H100 == hd2.H100 &&
		hd1.L == hd2.L &&
		hd1.Epsilon == hd2.Epsilon &&
		hd1.NSide == hd2.NSide &&
		hd1.NTotal == hd2.NTotal &&
		hd1.UniformMp == hd2.UniformMp 
}

func denseArrayEq(a1, a2 *container.DenseArray) bool {
	return a1.Length == a2.Length && a1.Bits == a2.Bits &&
		bytesEq(a1.Data, a2.Data)
}

func bytesEq(b1, b2 []byte) bool {
	if len(b1) != len(b2) { return false }
	for i := range b1 {
		if b1[i] != b2[i] { return false }
	}
	return true
}

func TestBlockIO(t *testing.T) {
	f, err := ioutil.TempFile(".", "test_*.lvec")
	if err != nil { panic(err.Error()) }
	defer os.Remove(f.Name())
	defer f.Close()

	hd := *testHeader

	rawHeader := []byte{1, 2, 3, 4, 5}
	hd.Offsets[1] = hd.Offsets[0] + 8 + uint64(len(rawHeader))

	vecs := make([]uint64, 3 * hd.SubCells*hd.SubCells*hd.SubCells)
	for i := range vecs { vecs[i] = uint64(i) + 10 }
	hd.SubCellVectorsMin, hd.SubCellVectorsBits = 10, 4
	vecArray := container.NewDenseArray(int(hd.SubCellVectorsBits), vecs)
	hd.Offsets[2] = hd.Offsets[1] + 8 + uint64(len(vecArray.Data))
	
	hd.BitsMin, hd.BitsBits = 6, 5
	nElem := uint64(hd.Hd.NSide) / (hd.Cells*hd.SubCells)
	nElem3 := nElem*nElem*nElem
	nSubCells3 := hd.SubCells*hd.SubCells*hd.SubCells

	bits := make([]uint64, 3 * nSubCells3)
	for i := range bits { bits[i] = uint64(i) + 6 }
	bitsArray := container.NewDenseArray(int(hd.BitsBits), bits)
	hd.Offsets[3] = hd.Offsets[2] + 8 + uint64(len(bitsArray.Data))

	arrays := make([]*container.DenseArray, 3*nSubCells3)
	k, totalLen := uint64(0), uint64(0)
	for i := range arrays {
		data := make([]uint64, nElem3)
		for j := range data {
			data[j] = k
			k++
		}

		arrays[i] = container.NewDenseArray(int(bits[i]), data)
		totalLen += uint64(len(arrays[i].Data))
	}

	hd.Offsets[4] = hd.Offsets[3] + 8 + totalLen

	err = writeHeaderBlock(f, &hd)
	if err != nil { t.Fatalf(err.Error()) }

	err = writeRawHeaderBlock(f, &hd, rawHeader)
	if err != nil { t.Fatalf(err.Error()) }

	err = writeSubCellVecsBlock(f, &hd, vecArray)
	if err != nil { t.Fatalf(err.Error()) }

	err = writeBitsBlock(f, &hd, bitsArray)
	if err != nil { t.Fatalf(err.Error()) }
	
	err = writeArraysBlock(f, &hd, arrays)
	if err != nil { t.Fatalf(err.Error()) }
	
	f.Seek(0, 0)
		
	rdHd, err := readHeaderBlock(f)
	if err != nil { t.Fatalf(err.Error()) }

	if !lvecHeaderEq(&hd, rdHd) {
		t.Errorf("written header, %v, not the same as the written header, %v",
			&hd, rdHd,
		)
	}

	rdRawHeader, err := readRawHeaderBlock(f, &hd) 
	if err != nil { t.Fatalf(err.Error()) }

	if !bytes.Equal(rdRawHeader, rawHeader) {
		t.Errorf("Wrote raw header %v, but got %v", rawHeader, rdRawHeader)
	}

	rdVecArray, err := readSubCellVecsBlock(f, &hd)
	if err != nil { t.Fatalf(err.Error()) }
	
	if !denseArrayEq(vecArray, rdVecArray) {
		t.Errorf("Wrote vector array %v, but read vector array %v.",
			vecArray, rdVecArray)
	}

	rdBitsArray, err := readBitsBlock(f, &hd)
	if err != nil { t.Fatalf(err.Error()) }

	if !denseArrayEq(bitsArray, rdBitsArray) {
		t.Errorf("Wrote vector array %v, but read vector array %v.",
			bitsArray, rdBitsArray)
	}

	rdArrays, err := readArraysBlock(f, &hd, bits)

	if len(arrays) != len(rdArrays) {
		t.Errorf("Length of arrays is %d, but length of read arrays is %d",
			len(arrays), len(rdArrays))
	} else {
		for i := range arrays {
			if !denseArrayEq(arrays[i], rdArrays[i]) {
				t.Errorf("array %d = %v, but read aray = %v",
					i, arrays[i], rdArrays[i])
			}
		}
	}
}

func uint64sEq(x1, x2 []uint64) bool {
	if len(x1) != len(x2) { return false }
	for i := range x1 {
		if x1[i] != x2[i] { return false }
	}
	return true
}

func TestBound(t *testing.T) {
	tests := []struct{
		x, xOut []uint64
		origin, width uint64
	} {
		{ []uint64{0}, []uint64{0}, 0, 1 },
		{ []uint64{0, 1}, []uint64{0, 1}, 0, 2 },
		{ []uint64{1, 2}, []uint64{0, 1}, 1, 2 },
		{ []uint64{10, 12, 14}, []uint64{0, 2, 4}, 10, 5 },
	}

	for i := range tests {
		x := make([]uint64, len(tests[i].x))
		for j := range x { x[j] = tests[i].x[j] }
		origin, width := bound(x)

		if origin != tests[i].origin {
			t.Errorf("test %d) expected origin = %d, got %d",
				i, tests[i].origin, origin)
		}
		if width != tests[i].width {
			t.Errorf("test %d) expected width = %d, got %d",
				i, tests[i].width, width)
		}
		if !uint64sEq(tests[i].xOut, x) {
			t.Errorf("test %d) expected x = %d, got %d",
				i, tests[i].xOut, x)
		}

		unbound(origin, x)

		if !uint64sEq(tests[i].x, x) {
			t.Errorf("test %d) bound(unbound) gave %d, not %d",
				i, x, tests[i].x)
		}
	}
}

func TestPeriodicBound(t *testing.T) {
	tests := []struct{
		x, xOut []uint64
		origin, width, pix uint64
	} {
		{ []uint64{0}, []uint64{0}, 0, 1, math.MaxInt64 },
		{ []uint64{0, 1}, []uint64{0, 1}, 0, 2, math.MaxInt64 },
		{ []uint64{1, 2}, []uint64{0, 1}, 1, 2, math.MaxInt64 },
		{ []uint64{10, 12, 14}, []uint64{0, 2, 4}, 10, 5, math.MaxInt64 },
		{ []uint64{10, 0, 2}, []uint64{0, 2, 4}, 10, 5, 12 },
		{ []uint64{10, 12, 14}, []uint64{0, 2, 4}, 10, 5, 15 },
		{ []uint64{1, 2, 3}, []uint64{1, 2, 3}, 0, 4, 4 }, 
		{ []uint64{8, 9, 0, 1}, []uint64{0, 1, 2, 3}, 8, 4, 10}, 
		{ []uint64{9, 0, 1, 8}, []uint64{1, 2, 3, 0}, 8, 4, 10}, 
		{ []uint64{0, 1, 8, 9}, []uint64{2, 3, 0, 1}, 8, 4, 10}, 
		{ []uint64{1, 8, 9, 0}, []uint64{3, 0, 1, 2}, 8, 4, 10}, 
	}

	for i := range tests {
		x := make([]uint64, len(tests[i].x))
		for j := range x { x[j] = tests[i].x[j] }
		origin, width := periodicBound(tests[i].pix, x)

		if origin != tests[i].origin {
			t.Errorf("test %d) expected origin = %d, got %d",
				i, tests[i].origin, origin)
		}
		if width != tests[i].width {
			t.Errorf("test %d) expected width = %d, got %d",
				i, tests[i].width, width)
		}
		if !uint64sEq(tests[i].xOut, x) {
			t.Errorf("test %d) expected x = %d, got %d",
				i, tests[i].xOut, x)
		}

		periodicUnbound(tests[i].pix, origin, x)

		if !uint64sEq(tests[i].x, x) {
			t.Errorf("test %d) bound(unbound) gave %d, not %d",
				i, x, tests[i].x)
		}
	}
}

func TestToArray(t *testing.T) {
	tests := []struct{
		x []uint64
		bits, min uint64
	}{
		{[]uint64{0}, 1, 0},
		{[]uint64{9}, 1, 9},
		{[]uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, 4, 0},
		{[]uint64{0, 1, 2, 3}, 3, 0},
		{[]uint64{4, 5, 6, 7}, 3, 4},
		{[]uint64{9, 8, 1, 0}, 4, 0},
	}

	for i := range tests {
		x := make([]uint64, len(tests[i].x))
		for j := range x { x[j] = tests[i].x[j] }

		bits, min, arr := toArray(x, 0, false)

		if bits != tests[i].bits {
			t.Errorf("test %d) expected bits = %d, got %d",
				i, tests[i].bits, bits)
		}
		if min != tests[i].min {
			t.Errorf("test %d) expected min = %d, got %d",
				i, tests[i].min, min)
		}

		buf := make([]uint64, len(x))
		loadArray(0, min, arr, buf)
		if !uint64sEq(buf, tests[i].x) {
			t.Errorf("test %d) expected x = %d, got %d", i, tests[i].x, buf)
		}
	} 
}


func TestPeriodicToArray(t *testing.T) {
	pix := uint64(10)
	tests := []struct{
		x []uint64
		bits, min uint64
	}{
		{[]uint64{0}, 1, 0},
		{[]uint64{9}, 1, 9},
		{[]uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, 4, 0},
		{[]uint64{0, 1, 2, 3}, 3, 0},
		{[]uint64{4, 5, 6, 7}, 3, 4},
		{[]uint64{9, 8, 1, 0}, 3, 8},
	}

	for i := range tests {
		x := make([]uint64, len(tests[i].x))
		for j := range x { x[j] = tests[i].x[j] }

		bits, min, arr := toArray(x, pix, true)

		if bits != tests[i].bits {
			t.Errorf("test %d) expected bits = %d, got %d",
				i, tests[i].bits, bits)
		}
		if min != tests[i].min {
			t.Errorf("test %d) expected min = %d, got %d",
				i, tests[i].min, min)
		}

		buf := make([]uint64, len(x))
		loadArray(pix, min, arr, buf)
		if !uint64sEq(buf, tests[i].x) {
			t.Errorf("test %d) expected x = %d, got %d", i, tests[i].x, buf)
		}
	} 
}

func TestDequantize(t *testing.T) {
	nElem3 := 100

	snap := &lvecSnapshot{
		quantBuf: make([]uint64, nElem3),
		hd: lvecHeader{
			Pix: 100,
			Limits: [2]float64{100, 150},
		},
	}

	for i := range snap.quantBuf { snap.quantBuf[i] = uint64(i) }
	out := make([][3]float32, nElem3)

	for dim := uint64(0); dim < 3; dim++ {
		snap.dequantize(out, dim)
	}

	for i := range out {
		for dim := 0; dim < 3; dim++ {		
			if out[i][dim] < 100 + float32(i)*0.5 ||
				out[i][dim] > 100 + float32(i+1)*0.5 {
				t.Fatalf("out[%d][%d] = %g", i, dim, out[i][dim])
			}
		}
	}
}

func TestLoadCell(t *testing.T) {
	snap := &lvecSnapshot{ }
	snap.hd.Pix = 125 * 8
	snap.hd.Cells = 3
	snap.hd.Hd.NSide = 30
	snap.hd.SubCells = 2
	snap.subCellBuf = make([]uint64, 125)
	snap.quantBuf = make([]uint64, 1000)
	
	arrs := make([]*container.DenseArray, 8)
	bits, mins := make([]uint64, 8), make([]uint64, 8)

	for i := range arrs {
		xoff := 5 * (i % 2)
		yoff := 5 * ((i / 2) % 2)
		zoff := 5 * (i / 4)

		j := 0
		x := make([]uint64, 125)
		for iz := zoff; iz < zoff + 5; iz++ {
			for iy := yoff; iy < yoff + 5; iy++ {
				for ix := xoff; ix < xoff + 5; ix++ {
					x[j] = uint64(ix + iy*10 + iz*100)
					j++
				}
			}
		}

		bits[i], mins[i], arrs[i] = toArray(x, snap.hd.Pix, false)
	}

	for i := 0; i < 8; i++ {
		snap.loadSubCell(uint64(i), mins[i], arrs[i])
	}


	for i := range snap.quantBuf { 
		if snap.quantBuf[i] != uint64(i) {
			t.Errorf("snap.quantBuf[%d] = %d", i, snap.quantBuf[i])
		}
	}
}

func TestMockLVec(t *testing.T) {
	hd := &Header{
		Z: 1.0, Scale: 0.5,
		OmegaM:0.3, OmegaL: 0.7, H100: 0.7,
		L: 10, Epsilon: 1,
		NSide: 10, NTotal: 1000,
		UniformMp: 1e10,
	}

	nSide3 := hd.NSide*hd.NSide*hd.NSide

	x := [][][3]float32{ make([][3]float32, nSide3) }
	v := [][][3]float32{ make([][3]float32, nSide3) }
	id := [][]int64{ make([]int64, nSide3) }

	i := 0
	for iz := 0; iz < int(hd.NSide); iz++ {
		for iy := 0; iy < int(hd.NSide); iy++ {
			for ix := 0; ix < int(hd.NSide); ix++ {
				x[0][i] = [3]float32{ float32(ix), float32(iy), float32(iz) }
				v[0][i] = [3]float32{ -float32(ix), float32(iy), -float32(iz) }
				id[0][i] = int64(i+1)
				i++
			}
		}
	}

	snap := NewMockSnapshot(hd, x, v, id)

	ConvertToLVec(snap, 1, 2, 0.1, 0.01, "test_lvec_data", "test.%s.%d.lvec")
	lvec, err := LVec("test_lvec_data", "test.%s.%d.lvec")
	if err != nil { panic(err.Error()) }

	idLvec, err := lvec.ReadID(0)
	if err != nil { panic(err.Error()) }
	for i := range idLvec {
		if idLvec[i] != int64(i) {
			t.Errorf("lvec.ID[%d] = %d", i, idLvec[i])
		}
	}

	mpLvec, err := lvec.ReadMp(0)
	if err != nil { panic(err.Error()) }
	for i := range mpLvec {
		if mpLvec[i] != float32(1e10) {
			t.Errorf("lvec.Mp[%d] = %g", i, mpLvec[i])
		}
	}

	xLvec, err := lvec.ReadX(0)
	if err != nil { panic(err.Error()) }
	i = 0
	for iz := 0; iz < 10; iz++ {
		for iy := 0; iy < 10; iy++ {
			for ix := 0; ix < 10; ix++ {
				vec := xLvec[i]
				if !floatEq(vec[0], float32(ix), 0.1) && 
					!floatEq(vec[1], float32(iy), 0.1) && 
					!floatEq(vec[2], float32(iz), 0.1) {
					t.Errorf("lvec.X[%d] = %f", i, vec)
				}
				i++
			}
		}
	}

	vLvec, err := lvec.ReadV(0)
	if err != nil { panic(err.Error()) }
	i = 0
	for iz := 0; iz < 10; iz++ {
		for iy := 0; iy < 10; iy++ {
			for ix := 0; ix < 10; ix++ {
				vec := vLvec[i]
				if !floatEq(vec[0], -float32(ix), 0.01) && 
					!floatEq(vec[1], float32(iy), 0.01) && 
					!floatEq(vec[2], -float32(iz), 0.01) {
					t.Errorf("lvec.V[%d] = %f", i, vec)
				}
				i++
			}
		}
	}
}

func floatEq(x, y, eps float32) bool {
	return x - eps < y && x + eps > y
}
