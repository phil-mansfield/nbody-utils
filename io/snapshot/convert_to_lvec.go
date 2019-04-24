package snapshot

import (
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
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

// VectorGrid is a segmented cubic grid that stores float32 vectors in cubic
// sub-segments.
type VectorGrid struct {
	Grid
	Cells [][][3]float32
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
	out := [3][]int64{}
	for i := 0; i < 3; i++ {
		out[i] = make([]int64, vg.NSide*vg.NSide*vg.NSide)
	}

	return out
}

// Quantize quantizes the cell, c, of a VectorGrid. The grid has a range given
// by and after quantization there should be pix "pixels" of resolutoin on
// one side. Each int64 slice in out must be of length vg.NSide^3.
func (vg *VectorGrid) Quantize(
	c int, pix int64, lim [2]float64, out [3][]uint64,
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
func PeriodicBound(pix int64, x []uint64) (origin, width uint64) {
	x0, width := x[0], int64(1)
	for _, xi := range x {
		x1 := x0 + width - 1
		if x1 >= pix { x1 -= pix }

		d0 := periodicDistance(xi, x0, pix)
		d1 := periodicDistance(xi, x1, pix)

		if d0 > 0 && d1 < 0 { continue }

		if d1 > -d0 {
			width += d1
		} else {
			x0 += d0
			if x0 < 0 { x0 += pix }
			width -= d0
		}

		if width > pix/2 { return 0, pix }
	}

	return x0, width
}

func periodicDistance(x, x0, pix uint64) uint64 {
	d := x - x0
	if d >= 0 {
		if d > pix - d { return d - pix }
	} else {
		if d < -(d + pix) { return pix + d }
	}
	return d
}

func Clip(pix, origin int64, x []int64) {
	for i := range x {
		x[i] -= origin
		if x[i] < 0 { x[i] += pix }
	}
}


func ConvertToLVec(
	snap Snapshot,
	cells, subCells int,
	dx, dv float64,
	dir, fnameFormat string,
	headerOption ...Header,
) error {
	// If the user wants to update the Header, use that instead.
	hd := snap.Header()
	if len(headerOption) { hd = headerOption[0] }

	if cells*subCells != hd.NSide {
		panic("cells = %d, subCells = %d, but hd.NSide = %d",
			cells, subCells, hd.NSide)
	} else if !snap.UniformMass() {
		panic("Currently, non-uniform masses aren't implemented.")
	} else if _, err := os.Stat(dir); os.IsNotExist(err) {
		panic(fmt.Sprintf("Driectory %s does not exist", dir))
	}

	lvHeader := &lvecHeader{
		magic: LVecMagicNumber,
		version: LVecVersion,
		method: lvecBoxMethod,
		hd: *hd
	}
	
	xGrid, xLim := XGrid(snap, cells*subCells), [2]float64{0, hd.L}
	lvHeader.varType = lvecX
	err := writeLvec(lvHeader, xGrid, xLim, dir, fnameFormat)
	if err != nil { return err }

	runtime.GC()

	vGrid, xLim := VGrid(snap, cells*subCells), [2]float64{-5000, 5000}
	lvHeader.varType = lvecX
	err = writeLvec(lvHeader, vGrid, vLim, dir, fnameFormat)
	if err != nil { return err }

	return nil
}

func writeLvec(
	hd *lvecHeader,
	grid *VectorGrid, lim [2]float64,
	dir, fnameFormat string
) err {
	nCells := hd.cells*hd.cells*hd.cells
	nSub := hd.subCells*hd.subCells*hd.subCells

	bits := make([][3]byte, nSub)
	offsets := make([][3]uint16, nSub)
	arrays := make([]*DenseArray, nSub)

	fnames := fnameList(hd, dir, fnameFormat)

	for c := 0; c < nCell; c++ {
		cx := c % hd.cells
		cy := (c / hd.cells) % hd.cells
		cz := c / (hd.cells * hd.cells)
		for s := 0; s < nSub; s++ {
			sx := s % hd.subCells
			sy := (s / hd.subCells) % hd.subCells
			sz := s / (hd.subCells * hd.subCells)

			ix := cx*hd.subCells + sx
			iy := cx*hd.subCells + sx
			iz := cx*hd.subCells + sx
		}
	}
}

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

func minPix(lim [2]float64, dx float64) int64 {
	return int64(math.Ceil((lim[1] - lim[0]) / dx))
}

/*
	dx := []float64{5.8e-3/3, 5.8e-3, 5.8e-3*3}
	dv := []float64{0.1, 0.3, 1.0}
	xRange := [2]float64{0, 62.5}
	vRange := [2]float64{-3500, 3500}
	
	xQuant := xGrid.IntBuffer()
	fmt.Println("Position size:")
	for i := range dx {
		pix := minPix(xRange, dx[i])

		size := 0

		for c := range xGrid.Cells {
			xGrid.Quantize(c, pix, xRange, xQuant)
			_, width0 := snapshot.PeriodicBound(pix, xQuant[0])
			_, width1 := snapshot.PeriodicBound(pix, xQuant[1])
			_, width2 := snapshot.PeriodicBound(pix, xQuant[2])
			bit := bits(width0)+bits(width1)+bits(width2)
			size += bit*len(xGrid.Cells[c]) + 16*3 + 8
		}
		fmt.Printf("%.3g ", float64(size / 8))
	}
	fmt.Println()

	vQuant := vGrid.IntBuffer()
	fmt.Println("Velocity size:")
	for i := range dx {
		pix := minPix(vRange, dv[i])

		size := 0

		for c := range vGrid.Cells {
			xGrid.Quantize(c, pix, vRange, vQuant)
			_, width0 := snapshot.PeriodicBound(pix, vQuant[0])
			_, width1 := snapshot.PeriodicBound(pix, vQuant[1])
			_, width2 := snapshot.PeriodicBound(pix, vQuant[2])
			bit := bits(width0)+bits(width1)+bits(width2)
			size += bit*len(vGrid.Cells[c]) + 16*3 + 8
		}
		fmt.Printf("%.3g ", float64(size / 8))
	}
*/
