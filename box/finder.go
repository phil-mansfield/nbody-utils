/*package box contains routines for dealing with the periodic geometry of
cosmological simulations boxes.*/
package box

const (
	defaultFinderCells = 250
)

// SplitSubhaloFinder finds the halos in group A that are within the radii of
// group B. It is optimized for the case where A is extremely large and search
// radii are also very large.
//
// To do this, it does not allow for host list identifications and does not
// memoize results.
type Finder struct {
	g      *Grid
	gBuf   []int
	idxBuf []int
	dr2Buf []float64
	x      [][3]float64
	bufi   int
}

// NewSubhaloFinder creates a new subhalo finder corresponding to the given
// Grid. The Grid contains halos from group A.
func NewFinder(L float64, x [][3]float64) *Finder {
	g := NewGrid(defaultFinderCells, L, len(x))
	
	f := &Finder{
		g: g,
		gBuf: make([]int, g.MaxLength()),
		idxBuf: make([]int, len(g.Next)),
		dr2Buf: make([]float64, len(g.Next)),
		x: x,
	}

	return f
}



// FindSubhalos links grid halos (from group A) to a target halo (from group B).
// Returned arrays are internal buffers, so please treat them kindly.
func (sf *Finder) Find(pos [3]float64, r0 float64) []int {
	sf.bufi = 0
	sf.idxBuf = sf.idxBuf[:cap(sf.idxBuf)]
	sf.dr2Buf = sf.dr2Buf[:cap(sf.dr2Buf)]

	b := &Bounds{}
	c := sf.g.Cells

	b.SphereBounds(pos, r0, sf.g.cw, sf.g.Width)

	for dz := 0; dz < b.Span[2]; dz++ {
		z := b.Origin[2] + dz
		if z >= c {
			z -= c
		}
		zOff := z * c * c
		for dy := 0; dy < b.Span[1]; dy++ {
			y := b.Origin[1] + dy
			if y >= c {
				y -= c
			}
			yOff := y * c
			for dx := 0; dx < b.Span[0]; dx++ {
				x := b.Origin[0] + dx
				if x >= c {
					x -= c
				}
				idx := zOff + yOff + x

				sf.gBuf = sf.g.ReadIndexes(idx, sf.gBuf)
				sf.addSubhalos(sf.gBuf, pos[0], pos[1], pos[2], r0, sf.g.Width)
			}
		}
	}

	return sf.idxBuf[:sf.bufi]
}

func (sf *Finder) addSubhalos(
	idxs []int, xh, yh, zh, rh float64, L float64,
) {
	for _, j := range idxs {
		sx, sy, sz := sf.x[j][0], sf.x[j][1], sf.x[j][2]
		dx, dy, dz, dr := xh-sx, yh-sy, zh-sz, rh

		if dx > +L/2 { dx -= L }
		if dx < -L/2 { dx += L }
		if dy > +L/2 { dy -= L }
		if dy < -L/2 { dy += L }
		if dz > +L/2 { dz -= L }
		if dz < -L/2 { dz += L }

		dr2 := dx*dx + dy*dy + dz*dz

		if dr*dr >= dr2 {
			sf.idxBuf[sf.bufi] = j
			sf.dr2Buf[sf.bufi] = dr2
			sf.bufi++
		}
	}
}
