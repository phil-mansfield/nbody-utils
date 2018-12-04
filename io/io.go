/*package io handles reading data files. I'll make this legit later on, but for
now it's just a loose wrapper around Shellfish's VectorBuffer struct.*/
package io

import (
	"fmt"
	"runtime"
	"github.com/phil-mansfield/nbody-utils/box"
	"github.com/phil-mansfield/nbody-utils/thread"
	sio "github.com/phil-mansfield/shellfish/io"
)

func NewFiles(buf sio.VectorBuffer, fnames []string) *Files {
	return &Files{ buf, fnames, nil, nil, nil, nil }
}

type Files struct {
	buf sio.VectorBuffer
	Names []string

	x, v [][3]float64
	m    []float64
	id   []int
}

func (f *Files) Len() int { return len(f.Names) }

func (f *Files) Read(i int, subsample ...int) (
	xp, vp [][3]float64, mp []float64, idp []int,
) {
	sub := 1
	if len(subsample) > 0 { sub = subsample[0] }
	
	runtime.GC()

	x, v, m, id, err := f.buf.Read(f.Names[i])
	defer f.buf.Close()
	if err != nil { panic(err.Error()) }

	n := len(x) / sub

	f.x, f.v = expandVec(f.x, n), expandVec(f.v, n)
	f.m, f.id =  expandFloat(f.m, n), expandInt(f.id, n)

	fmt.Println(len(f.x))
	
	for i := 0; i < len(f.x); i++ {
		for j := 0; j < 3; j++ {
			f.x[i][j] = float64(x[i*sub][j])
			f.v[i][j] = float64(v[i*sub][j])
		}
		f.m[i] = float64(m[i*sub])
		f.id[i] = int(id[i*sub])
	}

	return f.x, f.v, f.m, f.id
}

type HaloWork func(
	i int, // Halo index
	dx [][3]float64, dv [][3]float64, // relative to halo center
	m []float64, id []int,
)

func (f *Files) HaloLoop(
	hx, hv [][3]float64, hr []float64, work HaloWork,
	subsample ...int,
) {	
	for iFile := 0; iFile < f.Len(); iFile++ {
		f.fileLoop(iFile, hx, hv, hr, work)
	}
}

func (f *Files) fileLoop(
	iFile int,
	hx, hv [][3]float64, hr []float64, work HaloWork,
	subsample ...int,
) {
	if hv == nil { hv = make([][3]float64, len(hx)) }

	xp, vp, mp, idp := f.Read(iFile, subsample...)
	
	// Make finder; need the simulation width
	hd := &sio.Header{}
	finders := make([]*box.Finder, runtime.NumCPU())
	for i := range finders {
		f.buf.ReadHeader(f.Names[iFile], hd)
		finders[i] = box.NewFinder(hd.TotalWidth, xp)
	}

	xbuf, vbuf := [][3]float64{}, [][3]float64{}
	mbuf, idbuf := []float64{}, []int{}
	
	loopWork := func (worker, istart, iend, istep int) {
		for ih := istart; ih < iend; ih += istep { // halo index
			// Indices of particles in halo
			idx := finders[worker].Find(hx[ih], hr[ih])
			
			// resize buffers
			xbuf, vbuf = expandVec(xbuf, len(idx)), expandVec(vbuf, len(idx))
			mbuf = expandFloat(mbuf, len(idx))
			idbuf = expandInt(idbuf, len(idx))
			
			for j := range idx {
				ip := idx[j] // particle index
				for k := 0; k < 3; k ++ {
					xbuf[j][k] = box.SymBound(
						xp[ip][k] - hx[ih][k], hd.TotalWidth,
					)
					vbuf[j][k] = vp[ip][k] - hv[ih][k]
				}
				mbuf[j] = mp[ip]
				idbuf[j] = idp[ip]
			}

			// Finally, do the work:
			work(ih, xbuf, vbuf, mbuf, idbuf)
		}
	}
	
	thread.SplitArray(len(hx), runtime.NumCPU(), loopWork, thread.Jump())
}

func expandVec(buf [][3]float64, n int) [][3]float64 {
	if cap(buf) >= n {
		return buf[:n]
	} else {
		return append(buf[:cap(buf)], make([][3]float64, n - cap(buf))...)
	}
}

func expandFloat(buf []float64, n int) []float64 {
	if cap(buf) >= n {
		return buf[:n]
	} else {
		return append(buf[:cap(buf)], make([]float64, n - cap(buf))...)
	}
}

func expandInt(buf []int, n int) []int {
	if cap(buf) >= n {
		return buf[:n]
	} else {
		return append(buf[:cap(buf)], make([]int, n - cap(buf))...)
	}
}
