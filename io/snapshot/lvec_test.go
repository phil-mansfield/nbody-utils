package snapshot

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	
	"unsafe"
)

var testHeader = &lvecHeader{
	magic: LVecMagicNumber,
	version: LVecVersion,
	varType: lvecX,
	method: lvecBoxMethod,
	idx: 0,
	cells: 3,
	subCells: 2,
	subCellVectorsMin: 0,
	subCellVectorsBits: 0,
	bitsMin: 0,
	bitsBits: 0,
	pix: 100,
	limits: [2]float64{0, 50},
	delta: 1.0,
	offsets: [4]uint64{8 + uint64(unsafe.Sizeof(lvecHeader{})), 0, 0, 0},
	hd: Header{ },
}

func TestHeaderBlockIO(t *testing.T) {
	f, err := ioutil.TempFile(".", "test_*.lvec")
	if err != nil { panic(err.Error()) }
	defer os.Remove(f.Name())
	defer f.Close()

	fmt.Println(unsafe.Sizeof(lvecHeader{ }), testHeader.offsets)

	err = writeHeaderBlock(f, testHeader)
	if err != nil { t.Fatalf(err.Error()) }
	f.Seek(0, 0)

	rdHeader, err := readHeaderBlock(f)
	if err != nil { t.Fatalf(err.Error()) }

	fmt.Println(rdHeader)
}
