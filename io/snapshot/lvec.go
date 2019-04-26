package snapshot

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"runtime"

	"github.com/phil-mansfield/nbody-utils/container"

	"unsafe"
)

const (
	LVecMagicNumber = 0xbadf00d
	LVecVersion = 1

	lvecBoxMethod= iota

	lvecX = iota
	lvecV
)

type lvecHeader struct {
	Magic   uint64 // Magic number confirming that this file is a .lvec file
	Version uint64 // Version number of the code that generated this file.
	VarType uint64 // Flag showing the type of variable stored in the file.
	Method  uint64 // Flag showing the method used to store data within
	               // the file.

	Idx   uint64 // The index of the current file in the Lagrangian grid.
	Cells uint64 // The number of file-sized cells on one side of the
	             // Lagrangian grid.

	SubCells           uint64 // The number of sub-cells used inside this file.
	SubCellVectorsMin  uint64 // 
	SubCellVectorsBits uint64 // Number of bits used to represent the vectors to
	                          // each sub-cell.

	BitsMin  uint64 // Minimum bits value.
	BitsBits uint64 // Number of bits used to represent the number of bits in
	                // each cell.
	

	Pix        uint64 // Number of "pixels" on one side of the quantiation grid.
	Limits [2]float64 // The minimum and maximum value of each side of the box.
	Delta     float64 // The user-specified delta parameter. Each component of
                      // each vector will be stored to at least this accuracy.

	Offsets [4]uint64 // The offsets to the start of the sub-cell vector block,
	                  // the bits block, and the data block, respetively. This
	                  // isn't neccessary, but it sure is convenient.

	Hd Header // The header for the simulation.
}


func bound(x []uint64) (origin, width uint64) {
	min, max := x[0], x[0]
	for i := range x {
		if x[i] < min { min = x[i] }
		if x[i] > max { max = x[i] }
	}

	for i := range x { x[i] -= min }

	return min, max - min + 1
}

func unbound(origin uint64, x []uint64) {
	for i := range x { x[i] += origin }
}

// PeriodicBound returns the periodic bounds on the data contained in the array
// x with a total width of pix. pix must be less than MaxInt64
func periodicBound(pix uint64, x []uint64) (origin, width uint64) {	
	if pix > math.MaxInt64 {
		panic(fmt.Sprintf("pix = %d, but MaxInt64 = %d", pix, math.MaxInt64))
	}

	x0, iwidth, ipix := int64(x[0]), int64(1), int64(pix)

	for _, xu := range x {
		xi := int64(xu)

		x1 := x0 + iwidth - 1
		if x1 >= ipix { x1 -= ipix }

		d0 := periodicDistance(xi, x0, ipix)
		d1 := periodicDistance(xi, x1, ipix)

		if d0 > 0 && d1 < 0 { continue }

		if d1 > -d0 {
			iwidth += d1
		} else {
			x0 += d0
			if x0 < 0 { x0 += ipix }
			iwidth -= d0
		}

		if iwidth > ipix/2 { return 0, uint64(ipix) }
	}

	for i := range x {
		xi := int64(x[i]) - x0
		if xi < 0 { xi += int64(pix) }
		x[i] = uint64(xi)
	}

	return uint64(x0), uint64(iwidth)
}

// periodicDistance computes the distance from x0 to x.
func periodicDistance(x, x0, pix int64) int64 {
	ix, ix0, ipix := int64(x), int64(x0), int64(pix)

	d := ix - ix0
	if d >= 0 {
		if d > ipix - d { return d - ipix }
	} else {
		if d < -(d + ipix) { return ipix + d }
	}
	return d
}

func periodicUnbound(pix, origin uint64, x []uint64) {
	for i := range x {
		x[i] += origin
		if x[i] >= pix { x[i] -= pix }
	}
}

// ConvertToLVec converts a snapshot to a set of LVec files. cells is the number
// of file-sized cells on one side, subCells is the number of within-file cells
// on one side, dx and dy are the accuracy parameters for distance and velocity,
// respectively, dir is the directory that the files will be wirtten to, and
// fnameFormat is the printf format string used to generate file names and must
// contain a %s verb followed by a %d verb.
func ConvertToLVec(
	snap Snapshot,
	cells, subCells uint64,
	dx, dv float64,
	dir, fnameFormat string,
) error {
	hd := snap.Header()

	if int64(cells*subCells) != hd.NSide {
		panic(fmt.Sprintf("cells = %d, subCells = %d, but hd.NSide = %d",
			cells, subCells, hd.NSide))
	} else if !snap.UniformMass() {
		panic("Currently, non-uniform masses aren't implemented.")
	} else if _, err := os.Stat(dir); os.IsNotExist(err) {
		panic(fmt.Sprintf("Driectory %s does not exist", dir))
	}

	lvHeader := &lvecHeader{
		Magic: LVecMagicNumber,
		Version: LVecVersion,
		Cells: cells,
		SubCells: subCells,
		Method: lvecBoxMethod,
		Hd: *hd,
	}
	
	grid, err := XGrid(snap, int(cells*subCells))
	lvHeader.Limits = [2]float64{0, hd.L}
	lvHeader.VarType = lvecX
	lvHeader.Delta = dx
	lvHeader.Pix = minPix(lvHeader.Limits, lvHeader.Delta)

	err = generateLVec(lvHeader, grid, dir, fnameFormat)
	if err != nil { return err }

	runtime.GC()

	grid, err = VGrid(snap, int(cells*subCells))
	lvHeader.Limits = grid.Limits()
	lvHeader.VarType = lvecV
	lvHeader.Delta = dv
	lvHeader.Pix = minPix(lvHeader.Limits, lvHeader.Delta)

	err = generateLVec(lvHeader, grid, dir, fnameFormat)
	if err != nil { return err }

	return nil
}

// generateLVec generates all the LVec files associated with the data in grid.
func generateLVec(
	hd *lvecHeader,
	grid *VectorGrid,
	dir, fnameFormat string,
) error {
	cells, sCells := hd.Cells, hd.SubCells
	nSub := hd.SubCells*hd.SubCells*hd.SubCells

	subCellVecs := make([]uint64, 3*nSub)
	bits := make([]uint64, 3*nSub)
	arrays := make([]*container.DenseArray, 3*nSub)

	fnames := fnameList(hd, dir, fnameFormat)
	quant := grid.IntBuffer()

	grid.SuperCellLoop(cells, func(c uint64, cIdx [3]uint64) {
		runtime.GC()

		grid.SubCellLoop(cells, sCells, cIdx, func(i,s uint64, sIdx [3]uint64) {
			grid.Quantize(int(i), hd.Pix, hd.Limits, quant)
			
			for j := uint64(0); j < 3; j++ {
				k := 3*s + j
				bits[k], subCellVecs[k], arrays[k] = toArray(
					quant[j], hd.Pix, hd.VarType == lvecX,
				)
			}
		})

		var bitsArray, subCellVecsArray *container.DenseArray
		hd.BitsBits, hd.BitsMin, bitsArray = toArray(
			bits, 0, false,
		)
		hd.SubCellVectorsBits, hd.SubCellVectorsMin, subCellVecsArray = toArray(
			bits, 0, false,
		)

		err := writeLVecFile(fnames[c], hd, subCellVecsArray, bitsArray, arrays)
		if err != nil { panic(err.Error()) } // .___.
	})

	return nil
}

// toArray compresses an array of integers into a DenseArray and returns the
// information neccessary to decode this array, the minimum value of the
// data and the number of bits in the DenseArray.
func toArray(x []uint64, pix uint64, periodic bool) (
	bits, min uint64, array *container.DenseArray,
) {
	var width uint64
	if periodic {
		min, width = periodicBound(pix, x)
	} else {
		min, width = bound(x)
	}

	bits = minBits(width)
	array = container.NewDenseArray(int(bits), x)

	return bits, min, array
}

// loadArray loads the contents of array into the buffer buf. If pix is
// non-zero, the values are treated as periodic within the range pix. min is
// the offset value for the array.
func loadArray(
	pix, min uint64, array *container.DenseArray, buf []uint64,
) {
	array.Slice(buf)
	for i := range buf {
		buf[i] += min
		if buf[i] >= pix { buf[i] -= pix }
	}
}

// writeLVecFile writes an LVec file to disk
func writeLVecFile(
	fname string, hd *lvecHeader,
	subCellVecs *container.DenseArray,
	bits *container.DenseArray, 
	arrays []*container.DenseArray,
) error {
	f, err := os.Create(fname)
	defer f.Close()
	if err != nil { return err }

	totalArrayData := uint64(0)
	for _, a := range arrays { totalArrayData += uint64(len(a.Data)) }

	hd.Offsets[0] = uint64(unsafe.Sizeof(*hd)) + 8
	hd.Offsets[1] = uint64(len(subCellVecs.Data)) + 8 + hd.Offsets[0]
	hd.Offsets[2] = uint64(len(bits.Data)) + 8 + hd.Offsets[1]
	hd.Offsets[3] = uint64(totalArrayData) + 8 + hd.Offsets[2]
	fortranCheck(hd.Offsets)

	err = writeHeaderBlock(f, hd)
	if err != nil { return err }

	err = writeSubCellVecsBlock(f, hd, subCellVecs)
	if err != nil { return err }

	err = writeBitsBlock(f, hd, subCellVecs)
	if err != nil { return err }

	err = writeArraysBlock(f, hd, arrays)
	if err != nil { return err }

	return nil
}

// writeHeaderBlock writes the first LVec block, the header.
func writeHeaderBlock(f *os.File, hd *lvecHeader) error {
	offsetCheck(f, 0, "header", "start")

	err := binary.Write(f, binary.LittleEndian, int32(unsafe.Sizeof(*hd)))
	if err != nil { return err }
	err = binary.Write(f, binary.LittleEndian, hd)
	if err != nil { return err }
	err = binary.Write(f, binary.LittleEndian, int32(unsafe.Sizeof(*hd)))
	if err != nil { return err }

	return nil
}

// readHeaderBlock reads the first LVec bloc, the header.
func readHeaderBlock(f *os.File) (*lvecHeader, error) {
	offsetCheck(f, 0, "header", "start")

	fortran := [2]int32{ }
	hd := &lvecHeader{ }

	err := binary.Read(f, binary.LittleEndian, &fortran[0])
	if err != nil { return nil, err }
	err = binary.Read(f, binary.LittleEndian, hd)
	if err != nil { return nil, err }
	err = binary.Read(f, binary.LittleEndian, &fortran[1])
	if err != nil { return nil, err }

	fortranHeaderCheck(fortran, int(unsafe.Sizeof(*hd)), "header")
	offsetCheck(f, hd.Offsets[0], "header", "end")

	return hd, nil
}

// writeHeaderBlock writes the second LVec block, the quantized vectors to every
// sub-cell.
func writeSubCellVecsBlock(
	f *os.File, hd *lvecHeader, subCellVecs *container.DenseArray,
) error {
	offsetCheck(f, hd.Offsets[0], "vector", "start")

	err := binary.Write(f, binary.LittleEndian, int32(len(subCellVecs.Data)))
	if err != nil { return err }
	_, err = f.Write(subCellVecs.Data)
	if err != nil { return err }
	err = binary.Write(f, binary.LittleEndian, int32(len(subCellVecs.Data)))
	if err != nil { return err }

	offsetCheck(f, hd.Offsets[1], "vector", "end")

	return nil
}

// readSubCellVecsBlock reads the second LVec bloc, the vector block.
func readSubCellVecsBlock(
	f *os.File, hd *lvecHeader,
) (*container.DenseArray, error) {
	offsetCheck(f, hd.Offsets[0], "vector", "start")

	fortran := [2]int32{ }
	nSub := hd.SubCells*hd.SubCells*hd.SubCells
	array := createDenseArray(3*nSub, hd.SubCellVectorsBits)

	err := binary.Read(f, binary.LittleEndian, &fortran[0])
	if err != nil { return nil, err }
	_, err = io.ReadFull(f, array.Data)
	if err != nil { return nil, err }
	err = binary.Read(f, binary.LittleEndian, &fortran[1])
	if err != nil { return nil, err }

	fortranHeaderCheck(fortran, len(array.Data), "vector")
	offsetCheck(f, hd.Offsets[1], "vector", "end")

	return array, nil
}

// writeBitsBlock writes the third LVec block, the number of bits in each array.
func writeBitsBlock(
	f *os.File, hd *lvecHeader, bits *container.DenseArray,
) error {
	offsetCheck(f, hd.Offsets[1], "bits", "start")

	err := binary.Write(f, binary.LittleEndian, int32(len(bits.Data)))
	if err != nil { return err }
	_, err = f.Write(bits.Data)
	if err != nil { return err }
	err = binary.Write(f, binary.LittleEndian, int32(len(bits.Data)))
	if err != nil { return err }

	offsetCheck(f, hd.Offsets[2], "bits", "end")

	return nil
}

// readBitsBlock reads the third LVec bloc, the bits block.
func readBitsBlock(
	f *os.File, hd *lvecHeader,
) (*container.DenseArray, error) {
	offsetCheck(f, hd.Offsets[1], "bits", "start")

	fortran := [2]int32{ }
	nSub := hd.SubCells*hd.SubCells*hd.SubCells
	array := createDenseArray(3*nSub, hd.BitsBits)

	err := binary.Read(f, binary.LittleEndian, &fortran[0])
	if err != nil { return nil, err }
	_, err = io.ReadFull(f, array.Data)
	if err != nil { return nil, err }
	err = binary.Read(f, binary.LittleEndian, &fortran[1])
	if err != nil { return nil, err }

	fortranHeaderCheck(fortran, len(array.Data), "bits")
	offsetCheck(f, hd.Offsets[2], "bits", "end")

	return array, nil
}

// writeArraysBlock writes the fourth and final LVec block, the array data for
// each sub-cell.
func writeArraysBlock(
	f *os.File, hd *lvecHeader, arrays []*container.DenseArray,
) error {
	offsetCheck(f, hd.Offsets[2], "array", "start")

	totalArrayData := uint64(0)
	for _, a := range arrays { totalArrayData += uint64(len(a.Data)) }

	err := binary.Write(f, binary.LittleEndian, int32(totalArrayData))
	if err != nil { return err }
	for _, array := range arrays {
		_, err = f.Write(array.Data)
		if err != nil { return err }
	}
	err = binary.Write(f, binary.LittleEndian, int32(totalArrayData))
	if err != nil { return err }

	offsetCheck(f, hd.Offsets[3], "array", "end")

	return nil
}

// writeArraysBlock reads the fourth and final LVec block, the array data for
// each sub-cell.
func readArraysBlock(
	f *os.File, hd *lvecHeader, bits []uint64,
) ([]*container.DenseArray, error) {
	offsetCheck(f, hd.Offsets[2], "array", "start")

	fortran := [2]int32{ }

	err := binary.Read(f, binary.LittleEndian, &fortran[0])
	if err != nil { return nil, err }

	nCells := int(hd.SubCells*hd.SubCells*hd.SubCells)
	arrays := make([]*container.DenseArray, 3*nCells)
	nSubCellSide := uint64(hd.Hd.NSide) / (hd.Cells*hd.SubCells)
	for i := 0; i < 3*nCells; i++ {
		arrays[i] = createDenseArray(nSubCellSide, bits[i])
		_, err = io.ReadFull(f, arrays[i].Data)
		if err != nil { return nil, err }
	}

	err = binary.Read(f, binary.LittleEndian, &fortran[1])
	if err != nil { return nil, err }

	totalArrayData := 0
	for _, array := range arrays { totalArrayData += len(array.Data) }

	fortranHeaderCheck(fortran, totalArrayData, "array")
	offsetCheck(f, hd.Offsets[3], "array", "end")

	return arrays, nil
}

// fortranCheck ensures that all file blocks are small enough that they can have
// valid header/footer ints in Fortran. It panics if this is not true.
func fortranCheck(offsets [4]uint64) {
	headerSize := offsets[0] - 8
	subCellVecsSize := offsets[1] - offsets[0] - 8
	bitsSize := offsets[2] - offsets[1] - 8
	dataSize := offsets[3] - offsets[2] - 8
	
	if headerSize > math.MaxInt32 {
		panic(fmt.Sprintf("Internal failure: header block has size %d " + 
			"and will be too big to read by Fortran codes.", headerSize))
	} else if subCellVecsSize > math.MaxInt32 {
		panic(fmt.Sprintf("Internal failure: sub-cell vector block has size "+
			"%d and will be too big to read by Fortran codes.",subCellVecsSize))
	} else if bitsSize > math.MaxInt32 {
		panic(fmt.Sprintf("Internal failure: bits block has size %d " + 
			"and will be too big to read by Fortran codes.", bitsSize))
	} else if dataSize > math.MaxInt32 {
		panic(fmt.Sprintf("Internal failure: data block has size %d " + 
			"and will be too big to read by Fortran codes.", dataSize))
	}
}

// fortranHeaderCheck panics if the fortran header/footers are different
// sizes than the block that they enclose.
func fortranHeaderCheck(fortranHeaders [2]int32, n int, blockName string) {
	if fortranHeaders[0] != int32(n) || fortranHeaders[1] != int32(n) {
		panic(fmt.Sprintf("Internal I/O error: bits block has size %d, " + 
			"but the fortran header and footer are (%d, %d)",
			n, fortranHeaders[0], fortranHeaders[1]))
	}
}

// offsetCheck panics if a block does not start/end at the given offset from
// the start of the file.
func offsetCheck(f *os.File, offset uint64, blockName, startEnd string) {
	if loc, _ := f.Seek(0, 1); uint64(loc) != offset {
		panic(fmt.Sprintf("Internal I/O error: %s block %sed at byte " + 
			"%d, not byte %d.", blockName, startEnd, loc, offset))
	}
}

// createDenseArray creates and empty DenseArray.
func createDenseArray(nElem, bits uint64) *container.DenseArray {
	n := container.DenseArrayBytes(int(bits), int(nElem))
	return &container.DenseArray{
		Length: int(nElem), Bits: byte(bits), Data: make([]byte, n),
	}
}

// fnameList returns a slice of file names in dir that follow the format
// string fnameFormat corresponding to the data described by hd.
func fnameList(hd *lvecHeader, dir, fnameFormat string) []string {
	nCells := hd.Cells*hd.Cells*hd.Cells
	fnames := make([]string, nCells)

	for i := range fnames {
		typeName := ""
		switch hd.VarType {
		case lvecX: typeName = "X"
		case lvecV: typeName = "V"
		default:
			panic("Unrecognized lvecVarType")
		}
		
		fnames[i] = path.Join(dir, fmt.Sprintf(fnameFormat, typeName, i))
	}

	return fnames
}

// max returns the maximum element of x.
func max(x []uint64) uint64 {
	m := x[0]
	for i := range x {
		if m > x[i] { x[i] = m }
	}
	return m
}

// minPix returns the minimum number of pixels required to store points between
// [lim[0], lim[1]) with an accuracy of dx or better.
func minPix(lim [2]float64, dx float64) uint64 {
	return uint64(math.Ceil((lim[1] - lim[0]) / dx))
}

// minBits returns the number of bits needed to represent the number width.
func minBits(width uint64) uint64 {
	return uint64(math.Ceil(math.Log2(float64(width + 1))))
}
