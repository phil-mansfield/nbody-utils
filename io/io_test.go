package io

import (
	"testing"
	"github.com/phil-mansfield/nbody-utils/box"
	sio "github.com/phil-mansfield/shellfish/io"
)

func TestExpand(t *testing.T) {
	tests := []struct{ len, cap, newLen int } {
		{0, 0, 10},
		{0, 5, 10},
		{0, 5, 4},
		{0, 5, 5},
		{5, 5, 5},
		{4, 5, 4},
		{3, 5, 4},
		{4, 5, 3},
	}

	for i, test := range tests {
		xVec := make([][3]float64, test.len, test.cap)
		xFloat := make([]float64, test.len, test.cap)
		xInt := make([]int, test.len, test.cap)

		xVec = expandVec(xVec, test.newLen)
		xFloat = expandFloat(xFloat, test.newLen)
		xInt = expandInt(xInt, test.newLen)

		if len(xVec) != test.newLen {
			t.Errorf("Failed test %d. len(xVec) = %d.", i, len(xVec))
		}
		if len(xFloat) != test.newLen {
			t.Errorf("Failed test %d. len(xVec) = %d.", i, len(xFloat))
		}
		if len(xInt) != test.newLen {
			t.Errorf("Failed test %d. len(xVec) = %d.", i, len(xInt))
		}
	}
}

type FakeVectorBuffer struct {
	width float64
	xs, vs [][3]float32
	ms []float32
	ids []int64
}

	// Positions in Mpc/h and masses in Msun/h.
func (buf *FakeVectorBuffer) Read(fname string) (
	xs, vs [][3]float32, ms []float32, ids []int64, err error,
) {
	return buf.xs, buf.vs, buf.ms, buf.ids, nil
}
func (buf *FakeVectorBuffer) Close() {
}
func (buf *FakeVectorBuffer) IsOpen() bool {
	panic(":3")
}
func (buf *FakeVectorBuffer) ReadHeader(fname string, out *sio.Header) error {
	out.TotalWidth = buf.width
	return nil
}
func (buf *FakeVectorBuffer) MinMass() float32 {
	panic(":3")
}
func (buf *FakeVectorBuffer) TotalParticles(fname string) (int, error) {
	panic(":3")
}

func TestParticleCross(t * testing.T) {
	L := 10.0
	hx := [][3]float64{
		{5, 5, 5},
		{0, 7, 7},
		{7, 0, 7},
		{7, 7, 0},
		{10, 3, 3},
		{3, 10, 3},
		{3, 3, 10},
	}
	hv := make([][3]float64, len(hx))
	hr := make([]float64, len(hx))
	for i := range hv {
		hv[i] = [3]float64{1, 2, 3}
		hr[i] = 1.5
	}

	xp := make([][3]float32, len(hx) * 7)
	for i := range hx {
		for j := 0; j < 7; j++ {
			switch j {
			case 0:
				xp[i*7 + j] = vec64to32(hx[i], L)
			case 1:
				xp[i*7 + j] = vec64to32(
					[3]float64{hx[i][0]+1, hx[i][1], hx[i][2]}, L)
			case 2:
				xp[i*7 + j] = vec64to32(
					[3]float64{hx[i][0]-1, hx[i][1], hx[i][2]}, L)
			case 3:
				xp[i*7 + j] = vec64to32(
					[3]float64{hx[i][0], hx[i][1]+1, hx[i][2]}, L)
			case 4:
				xp[i*7 + j] = vec64to32(
					[3]float64{hx[i][0], hx[i][1]-1, hx[i][2]}, L)
			case 5:
				xp[i*7 + j] = vec64to32(
					[3]float64{hx[i][0], hx[i][1], hx[i][2]+1}, L)
			case 6:
				xp[i*7 + j] = vec64to32(
					[3]float64{hx[i][0], hx[i][1], hx[i][2]-1}, L)
			}
		}
	}

	vp := make([][3]float32, len(xp))
	m, id := make([]float32, len(xp)), make([]int64, len(xp))
	for i := range m {
		m[i], id[i] = float32(i), int64(i)
	}

	buf := &FakeVectorBuffer{L, xp, vp, m, id}
	files := Files{buf, []string{":3"}, nil, nil, nil, nil}

	files.HaloLoop(
		hx, hv, hr, func(i int, dx, dv [][3]float64, mp []float64, id []int) {
			if len(dx) != 7 || len(dv) != 7 || len(mp) != 7 || len(id) != 7 {
				t.Errorf("%d) |hx| = %d, |hv| = %d, |m| = %d |id| = %d",
					i, len(hx), len(hv), len(mp), len(id))
			}
		},
	)
}

func vec64to32(vec64 [3]float64, L float64) [3]float32 {
	return [3]float32{
		float32(box.Bound(vec64[0], L)),
		float32(box.Bound(vec64[1], L)),
		float32(box.Bound(vec64[2], L)),
	}
}
