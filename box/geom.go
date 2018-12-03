package box


// Bounds is a cell-aligned bounding box.
type Bounds struct {
	Origin, Span [3]int
}

// SphereBounds creates a cell-aligned bounding box around a non-aligned
// sphere within a box with periodic boundary conditions.
func (b *Bounds) SphereBounds(pos [3]float64, r, cw, width float64) {
	for i := 0; i < 3; i++ {
		min, max := pos[i]-r, pos[i]+r
		if min < 0 {
			min += width
			max += width
		}
		
		minCell, maxCell := int(min/cw), int(max/cw)
		b.Origin[i] = minCell
		b.Span[i] = maxCell - minCell + 1
	}
}

// ConvertIndices converts non-periodic indices to periodic indices.
func (b *Bounds) ConvertIndices(x, y, z, width int) (bx, by, bz int) {
	bx = x - b.Origin[0]
	if bx < 0 {
		bx += width
	}
	by = y - b.Origin[1]
	if by < 0 {
		by += width
	}
	bz = z - b.Origin[2]
	if bz < 0 {
		bz += width
	}
	return bx, by, bz
}

// Inside returns true if the given value is within the bounding box along the
// given dimension. The periodic box width is given by width.
func (b *Bounds) Inside(val int, width int, dim int) bool {
	lo, hi := b.Origin[dim], b.Origin[dim]+b.Span[dim]
	if val >= hi {
		val -= width
	} else if val < lo {
		val += width
	}
	return val < hi && val >= lo
}
