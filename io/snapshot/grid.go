package snapshot

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"math"
	"runtime"

	"github.com/phil-mansfield/nbody-utils/container"

	"unsafe"
)

// Grid manages the geometry of a cube which has been split up into cubic
// segments. It should be embedded in a struct which contains cube data of a
// specific type.
type Grid struct {
	NCell int64 // Number of cells on one side of the grid.
	NSide int64 // Number of elements on one side of a cell.
}


// Index returns the locaiton of a given ID in a Grid. Here, the id is assumed
// to be in the form of ix + iy*cells + iz*cells^2, and the location is
// specified by two indices, c and i. c is the index of the cell that the ID
// is within, and i is the index within that cell.
func (g *Grid) Index(id int64) (c, i int64) {
	nAll := g.NCell * g.NSide

	if id < 0 || id >= nAll*nAll*nAll {
		panic(fmt.Sprintf("ID %d is not valid for NCell = %d, NSide = %d",
			id, g.NCell, g.NSide))
	}

	idx := id % nAll
	idy := (id / nAll) % nAll
	idz := id / (nAll * nAll)

	ix, iy, iz := idx % g.NSide, idy % g.NSide, idz % g.NSide
	i = ix + iy*g.NSide + iz*g.NSide*g.NSide

	cx, cy, cz := idx / g.NSide, idy / g.NSide, idz / g.NSide
	c = cx + cy*g.NCell + cz*g.NCell*g.NCell

	return c, i
}

func superCellLoop(
	superCells uint64, 
	callback func(c uint64, cIdx [3]uint64),
) {
	nSuperCells := superCells * superCells * superCells

	for c := uint64(0); c < nSuperCells; c++ {
		runtime.GC()

		cx := c % superCells
		cy := (c / superCells) % superCells
		cz := c / (superCells * superCells)

		callback(c, [3]uint64{cx, cy, cz})
	}
}

func subCellLoop(
	superCells, subCells uint64, cIdx [3]uint64,
	callback func(i, s uint64, sIdx [3]uint64),
) {
	cx, cy, cz := cIdx[0], cIdx[1], cIdx[2]
	nSub := subCells * subCells * subCells
	cells := subCells*superCells

	for s := uint64(0); s < nSub; s++ {
		sx := s % subCells
		sy := (s / subCells) % subCells
		sz := s / (subCells * subCells)
		
		ix := cx*subCells + sx
		iy := cy*subCells + sy
		iz := cz*subCells + sz
		
		i := ix + iy*cells + iz*cells*cells

		callback(i, s, [3]uint64{sx, sy, sz})
	}
}

// VectorGrid is a segmented cubic grid that stores float32 vectors in cubic
// sub-segments.
type VectorGrid struct {
	Grid
	Cells [][][3]float32
}

// Limits returns the minimum and maximum values taken on by any component of
// any vector in the grid.
func (g *VectorGrid) Limits() [2]float64 {
	min, max := g.Cells[0][0][0], g.Cells[0][0][0]
	for _, cell := range g.Cells {
		for _, v := range cell {
			for _, x := range v {
				if x < min { min = x }
				if x > max { max = x }
			}
		}
	}

	return [2]float64{ float64(min), float64(max) }
}

// NewVectorGrid creates a new VectorGrid made with the specified total side
// length and number of cells on one side. cells must cleanly divide nSideTot.
func NewVectorGrid(cells, nSideTot int) *VectorGrid {
	nSide := nSideTot / cells
	if nSide*cells != nSideTot {
		panic(fmt.Sprintf("cells = %d doesn't evenly divide nSideTot = %d.",
			cells, nSideTot))
	}

	vg := &VectorGrid{
		Cells: make([][][3]float32, cells*cells*cells),
	}
	vg.Grid = Grid{NCell: int64(cells), NSide: int64(nSide)}

	for i := range vg.Cells {
		vg.Cells[i] = make([][3]float32, nSide*nSide*nSide)
	}

	return vg
}

// XGrid creates a VectorGrid of the position vectors in a snapshot.
func XGrid(snap Snapshot, cells int) (*VectorGrid, error) {
	hd := snap.Header()
	files := snap.Files()

	grid := NewVectorGrid(cells, int(hd.NSide))

	for i := 0; i < files; i++ {
		runtime.GC()

		x, err := snap.ReadX(i)
		if err != nil { return nil, err }
		id, err := snap.ReadID(i)
		if err != nil { return nil, err }
		for j := range x { grid.Insert(id[j] - 1, x[j]) }
	}

	return grid, nil
}

// VGrid creates a VectorGrid of the velocity vectors in a snapshot.
func VGrid(snap Snapshot, cells int) (*VectorGrid, error) {
	hd := snap.Header()
	files := snap.Files()

	grid := NewVectorGrid(cells, int(hd.NSide))

	for i := 0; i < files; i++ {
		runtime.GC()

		v, err := snap.ReadV(i)
		if err != nil { return nil, err }
		id, err := snap.ReadID(i)
		if err != nil { return nil, err }
		for j := range v { grid.Insert(id[j] - 1, v[j]) }
	}

	return grid, nil
}

// Insert inserts a vector into a VectorGrid.
func (vg *VectorGrid) Insert(id int64, v [3]float32) {
	c, i := vg.Index(id)
	vg.Cells[c][i] = v
}

func (vg *VectorGrid) IntBuffer() [3][]uint64 {
	out := [3][]uint64{}
	for i := 0; i < 3; i++ {
		out[i] = make([]uint64, vg.NSide*vg.NSide*vg.NSide)
	}

	return out
}

// Quantize quantizes the cell, c, of a VectorGrid. The grid has a range given
// by and after quantization there should be pix "pixels" of resolutoin on
// one side. Each int64 slice in out must be of length vg.NSide^3.
func (vg *VectorGrid) Quantize(
	c int, pix uint64, lim [2]float64, out [3][]uint64,
) {
	for i := 0; i < 3; i++ {
		if len(out[i]) != int(vg.NSide*vg.NSide*vg.NSide) {
			panic(fmt.Sprintf("len(out[%d]) = %d, but vg.NSide = %d.",
				i, len(out[i]), vg.NSide))
		}
	}

	L := lim[1] - lim[0]
	dx := float32(L / float64(pix))
	low := float32(lim[0])

	for i, v := range vg.Cells[c] {
		for j := 0; j < 3; j++ {
			out[j][i] = uint64((v[j] - low) / dx)
			// The next two lines should never be true unless there's some
			// floating point fuzziness.
			if out[j][i] < 0 { out[j][i] = 0 }
			if out[j][i] >= pix { out[j][i] = pix - 1 }
		}
	}
}

func Bound(x []uint64) (origin, width uint64) {
	min, max := x[0], x[0]
	for i := range x {
		if x[i] < min { min = x[i] }
		if x[i] > max { max = x[i] }
	}

	return min, max - min
}

// PeriodicBound returns the periodic bounds on the data contained in the array
// x with a total width of pix.
func PeriodicBound(pix uint64, x []uint64) (origin, width uint64) {
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

// Clip 
func Clip(pix, origin uint64, x []uint64) {
	for i := range x {
		xi := int64(x[i]) - int64(origin)
		if x[i] < 0 { xi += int64(pix) }
		x[i] = uint64(x[i])
	}
}


func ConvertToLVec(
	snap Snapshot,
	cells, subCells uint64,
	dx, dv float64,
	dir, fnameFormat string,
	headerOption ...Header,
) error {
	// If the user wants to update the Header, use that instead.
	hd := snap.Header()
	if len(headerOption) > 0 { hd = &headerOption[0] }

	if int64(cells*subCells) != hd.NSide {
		panic(fmt.Sprintf("cells = %d, subCells = %d, but hd.NSide = %d",
			cells, subCells, hd.NSide))
	} else if !snap.UniformMass() {
		panic("Currently, non-uniform masses aren't implemented.")
	} else if _, err := os.Stat(dir); os.IsNotExist(err) {
		panic(fmt.Sprintf("Driectory %s does not exist", dir))
	}

	lvHeader := &lvecHeader{
		magic: LVecMagicNumber, version: LVecVersion,
		cells: cells, subCells: subCells,
		method: lvecBoxMethod, hd: *hd,
	}
	
	grid, err := XGrid(snap, int(cells*subCells))
	lvHeader.limits = [2]float64{0, hd.L}
	lvHeader.varType = lvecX
	lvHeader.delta = dx
	lvHeader.pix = minPix(lvHeader.limits, lvHeader.delta)

	err = createLVec(lvHeader, grid, dir, fnameFormat)
	if err != nil { return err }

	runtime.GC()

	grid, err = VGrid(snap, int(cells*subCells))
	lvHeader.limits = grid.Limits()
	lvHeader.varType = lvecV
	lvHeader.delta = dv
	lvHeader.pix = minPix(lvHeader.limits, lvHeader.delta)

	err = createLVec(lvHeader, grid, dir, fnameFormat)
	if err != nil { return err }

	return nil
}

func createLVec(
	hd *lvecHeader,
	grid *VectorGrid,
	dir, fnameFormat string,
) error {
	nCells := hd.cells*hd.cells*hd.cells
	nSub := hd.subCells*hd.subCells*hd.subCells

	subCellVecs := make([]uint64, 3*nSub)
	bits := make([]uint64, 3*nSub)
	arrays := make([]*container.DenseArray, 3*nSub)

	fnames := fnameList(hd, dir, fnameFormat)
	quant := grid.IntBuffer()

	for c := uint64(0); c < nCells; c++ {
		runtime.GC()

		cx := c % hd.cells
		cy := (c / hd.cells) % hd.cells
		cz := c / (hd.cells * hd.cells)
		for s := uint64(0); s < nSub; s++ {
			sx := s % hd.subCells
			sy := (s / hd.subCells) % hd.subCells
			sz := s / (hd.subCells * hd.subCells)

			ix := cx*hd.subCells + sx
			iy := cy*hd.subCells + sy
			iz := cz*hd.subCells + sz

			i := ix + iy*uint64(grid.NCell) + iz*uint64(grid.NCell*grid.NCell)

			grid.Quantize(int(i), hd.pix, hd.limits, quant)

			for j := uint64(0); j < 3; j++ {
				var width uint64
				if hd.varType == lvecX {
					subCellVecs[3*s+j], width = PeriodicBound(hd.pix, quant[j])
				} else {
					subCellVecs[3*s+j], width = Bound(quant[j])
				}

				Clip(hd.pix, subCellVecs[3*s+j], quant[j])
				bits[3*s + j] = minBits(width)
			}
		}

		bitsMin, bitsWidth := Bound(bits)
		bitsBits := minBits(bitsWidth)
		subCellVecsMin, subCellVecsWidth := Bound(subCellVecs)
		subCellVecsBits := minBits(subCellVecsWidth)
		
		bitsArray := container.NewDenseArray(int(bitsBits), bits)
		subCellVecsArray := container.NewDenseArray(
			int(subCellVecsBits), subCellVecs,
		)
		
		hd.bitsMin,   hd.subCellVectorsMin = bitsMin,  subCellVecsMin
		hd.bitsBits, hd.subCellVectorsBits = bitsBits, subCellVecsBits

		err := writeLVecFile(fnames[c], hd, subCellVecsArray, bitsArray, arrays)
		if err != nil { return err }
	}

	return nil
}

func writeLVecFile(
	fname string, hd *lvecHeader,
	subCellVecs *container.DenseArray,
	bits *container.DenseArray, 
	arrays []*container.DenseArray,
) error {
	f, err := os.Create(fname)
	defer f.Close()
	if err != nil { return err }

	totalArrayData := 0
	for _, array := range arrays {
		totalArrayData += len(array.Data)
	}

	hd.offsets[0] = uint64(unsafe.Sizeof(*hd)) + 8
	hd.offsets[1] = uint64(len(subCellVecs.Data)) + 8 + hd.offsets[0]
	hd.offsets[2] = uint64(len(bits.Data)) + 8 + hd.offsets[1]
	hd.offsets[3] = uint64(totalArrayData) + 8 + hd.offsets[2]
	fortranCheck(hd.offsets)

	// header block
	err = binary.Write(f, binary.LittleEndian, uint64(unsafe.Sizeof(*hd)))
	if err != nil { return err }
	err = binary.Write(f, binary.LittleEndian, hd)
	if err != nil { return err }
	err = binary.Write(f, binary.LittleEndian, uint64(unsafe.Sizeof(*hd)))
	if err != nil { return err }

	// sub-cell vectors block
	err = binary.Write(f, binary.LittleEndian, uint64(len(subCellVecs.Data)))
	if err != nil { return err }
	err = binary.Write(f, binary.LittleEndian, subCellVecs.Data)
	if err != nil { return err }
	err = binary.Write(f, binary.LittleEndian, uint64(len(subCellVecs.Data)))
	if err != nil { return err }

	// bits block
	err = binary.Write(f, binary.LittleEndian, uint64(len(bits.Data)))
	if err != nil { return err }
	err = binary.Write(f, binary.LittleEndian, bits.Data)
	if err != nil { return err }
	err = binary.Write(f, binary.LittleEndian, uint64(len(bits.Data)))
	if err != nil { return err }

	// data block
	err = binary.Write(f, binary.LittleEndian, uint64(totalArrayData))
	if err != nil { return err }
	for _, array := range arrays {
		err = binary.Write(f, binary.LittleEndian, array.Data)
		if err != nil { return err }
	}
	err = binary.Write(f, binary.LittleEndian, uint64(totalArrayData))
	if err != nil { return err }

	return nil
}

// fortranCheck ensures that all file blocks are small enough that they can have
// valid header/footer ints in Fortran. It panics if this is not true.
//
// I'm super super happy that I still have to be doing this in
// the-year-of-our-lord-2019.
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

// fnameList returns a slice of file names in dir that follow the format
// string fnameFormat corresponding to the data described by hd.
func fnameList(hd *lvecHeader, dir, fnameFormat string) []string {
	nCells := hd.cells*hd.cells*hd.cells
	fnames := make([]string, nCells)

	for i := range fnames {
		typeName := ""
		switch hd.varType {
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
