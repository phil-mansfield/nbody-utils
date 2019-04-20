package snapshot

import (
	"github.com/phil-mansfield/nbody-utils/cosmo"
)

type Snapshot interface {
	Files() int // Number of files in the snapshot
	Header() *Header // Header contains basic information about the snapshot
	Index() *Index // Index to help with searching and snapshot intersection

	// All these methods return internal buffers, so don't append to them or
	// expect them to stick around after the function is called again.
	ReadX(i int) ([][3]float32, error) // Read positions for file i.
	ReadV(i int) ([][3]float32, error) // Read velocities for file i.
	ReadID(i int) ([]int64, error) // Read IDs for file i.
	ReadMp(i int) ([]float32, error) // Read particle masses for file i.
}

// Header is a struct containing basic information about the snapshot. Not all
// simulation headers provide all information: the user is responsible for
// supplying that information afterwards in these cases.
type Header struct {
	Z, Scale float64 // Redshift, scale factor
	OmegaM, OmegaL, H100 float64 // Omega_m(z=0), Omega_L(z=0), little-h(z=0)
	L, Epsilon float64 // Box size, force softening
	NSide, NTotal int64 // Particles on one size, total particles
	UniformMp float64 // If all particle masses are the same, this is m_p.
}


func (hd *Header) calcUniformMass() {
	rhoM0 := cosmo.RhoAverage(hd.H100*100, hd.OmegaM, hd.OmegaL, 0)
	mTot := (hd.L * hd.L * hd.L) * rhoM0
	hd.UniformMp =  mTot / float64(hd.NTotal)
}
