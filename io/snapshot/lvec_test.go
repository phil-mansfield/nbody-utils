package snapshot

import (
	"io/ioutil"
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
	SubCells: 2,
	SubCellVectorsMin: 0,
	SubCellVectorsBits: 0,
	BitsMin: 0,
	BitsBits: 0,
	Pix: 100,
	Limits: [2]float64{0, 50},
	Delta: 1.0,
	Offsets: [4]uint64{8 + uint64(unsafe.Sizeof(lvecHeader{})), 0, 0, 0},
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
	vecs := make([]uint64, 3 * hd.SubCells*hd.SubCells*hd.SubCells)
	for i := range vecs { vecs[i] = uint64(i) + 10 }
	hd.SubCellVectorsMin, hd.SubCellVectorsBits = 10, 4
	vecArray := container.NewDenseArray(int(hd.SubCellVectorsBits), vecs)
	hd.Offsets[1] = hd.Offsets[0] + 8 + uint64(len(vecArray.Data))
	
	hd.BitsMin, hd.BitsBits = 6, 5
	nElem := uint64(hd.Hd.NSide) / (hd.Cells*hd.SubCells)
	nElem3 := nElem*nElem*nElem
	nSubCells3 := hd.SubCells*hd.SubCells*hd.SubCells

	bits := make([]uint64, 3 * nSubCells3)
	for i := range bits { bits[i] = uint64(i) + 6 }
	bitsArray := container.NewDenseArray(int(hd.BitsBits), bits)
	hd.Offsets[2] = hd.Offsets[1] + 8 + uint64(len(bitsArray.Data))

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

	hd.Offsets[3] = hd.Offsets[2] + 8 + totalLen

	err = writeHeaderBlock(f, &hd)
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
		t.Errorf("Length of arrays is %d, but lenght of read arrays is %d",
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
