package snapshot

// Index contains boundary information for all files in a snapshot and allows
// for easily associating files with a particular region. It is created by
// calling Snapshot.Index()
type Index struct {
	// TODO
}

// Files returns the number of files int he snapshot.
func (idx *Index) Files() int {
	panic("NYI")
}

// Bounds returns the bounding box for file i.
func (idx *Index) Bounds(i int) *Bounds {
	panic("NYI")
}

// Find returns the indices of all the files that intersect with a given set of
// bounds.
func (idx *Index) Find(b *Bounds) []int {
	panic("NYI")
}

// Intersect returns the indices of every bound in a given set that intersect
// with the specified file.
func (idx *Index) Intersect(i int, b []Bounds) []int {
	panic("NYI")
}

// Bounds is a bounding box in 3D periodic space. It's advised that you create
// these through the functions BoundingBox and BoundingSphere
type Bounds struct {
	L float64
	Origin, Width [3]float64
}

// Buffer expands a bounding box to include a buffer region of width dx.
func (b *Bounds) Buffer(dx float64) *Bounds {
	panic("NYI")
}

// Intersect returns true if two bounding boxes intersect and false otherwise.
func Intersect(b1, b2 *Bounds) bool {
	panic("NYI")
}

// BoundingSphere returns the Bounds associated with a given sphere.
func BoundingSphere(Center [3]float64, Radius, L float64) *Bounds {
	panic("NYI")
}

// BoundingBox returns the Bounds associated with a given box.
func BoundingBox(Origin, Width [3]float64) *Bounds {
	panic("NYI")
}
