package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/phil-mansfield/nbody-utils/box"
	"github.com/phil-mansfield/nbody-utils/config"
	"github.com/phil-mansfield/nbody-utils/cosmo"
	"github.com/phil-mansfield/nbody-utils/io/catalogue"
)

type SimulationInfo struct {
	L float64
	IDIndex int64
	MassIndex int64
	PositionIndex []int64
	MassDef string
	RadiusMult float64

	OmegaM float64
	H0 float64
	ScaleFactor float64
}

func ReadSimulationInfo(fname string) *SimulationInfo {
	info := &SimulationInfo{ }

	vars := config.NewConfigVars("simulation_info")
	vars.Float(&info.L, "L", -1)
	vars.Float(&info.ScaleFactor, "ScaleFactor", -1)
	vars.Float(&info.OmegaM, "OmegaM", -1)
	vars.Float(&info.H0, "H0", -1)
	vars.Float(&info.RadiusMult, "RadiusMult", -1)
	vars.Int(&info.MassIndex, "MassIndex", -1)
	vars.Int(&info.MassIndex, "IDIndex", -1)
	vars.Ints(&info.PositionIndex, "PositionIndex", []int64{})
	vars.String(&info.MassDef, "MassDefinition", "")

	if err := config.ReadConfig(fname, vars); err != nil { panic(err.Error()) }

	if info.L < 0 {
		log.Fatalf("Must set L in %s", fname)
	} else if info.OmegaM < 0 {
		log.Fatalf("Must set OmegaM in %s", fname)
	} else if info.H0 < 0 {
		log.Fatalf("Must set H0 in %s", fname)
	} else if info.ScaleFactor < 0 {
		log.Fatalf("Must set ScaleFactor in %s", fname)
	} else if info.IDIndex < 0 {
		log.Fatalf("Must set IDIndex in %s", fname)
	} else if info.MassIndex < 0 {
		log.Fatalf("Must set MassIndex in %s", fname)
	} else if len(info.PositionIndex) == 0 {
		log.Fatalf("Must set PositionIndex in %s", fname)
	} else if info.MassDef == "" {
		log.Fatalf("Must set MassDefinition in %s", fname)
	} else if info.RadiusMult <= 0 {
		log.Fatalf("Must set RadiusMult in %s", fname)
	} 

	end := len(info.MassDef) - 1
	if len(info.PositionIndex) != 3 {
		log.Fatalf("Must give three MassIndex values in %s", fname)
	} else if info.MassDef != "vir" &&
		info.MassDef[end] != 'c' &&
		info.MassDef[end] != 'm' {
		log.Fatalf("Must set MassDefinition to 'vir', '___c', or '___m'")
	}

	return info
}

func RhoVir(info *SimulationInfo) float64 {
	a, omegaM, H0 := info.ScaleFactor, info.OmegaM, info.H0
	omegaL := 1 - omegaM
	z := 1/a - 1

	Hz := cosmo.HubbleFrac(omegaM, omegaL, z)
	omegaMz := (omegaM / (a*a*a)) / (Hz*Hz)
	x := omegaMz - 1

	deltaVir := (18*math.Pow(math.Pi, 2) + 82*x - 39*x*x) / omegaMz
	rhoVir := deltaVir * cosmo.RhoAverage(H0, omegaM, omegaL, z)

	return rhoVir
}

func RhoC(D float64, info *SimulationInfo) float64 {
	a, omegaM, H0 := info.ScaleFactor, info.OmegaM, info.H0
	omegaL := 1 - omegaM
	z := 1/a - 1
	return D * cosmo.RhoCritical(H0, omegaM, omegaL, z)
}

func RhoM(D float64, info *SimulationInfo) float64 {
	a, omegaM, H0 := info.ScaleFactor, info.OmegaM, info.H0
	omegaL := 1 - omegaM
	z := 1/a - 1
	return D * cosmo.RhoAverage(H0, omegaM, omegaL, z)
}

func Radius(mass []float64, info *SimulationInfo) []float64 {
	radius := make([]float64, len(mass))

	var rhoVir float64
	end := len(info.MassDef) - 1
	if info.MassDef == "vir" {
		rhoVir = RhoVir(info)
	} else if info.MassDef[end] == 'c' {
		D, err := strconv.ParseFloat(info.MassDef[:end], 64)
		if err != nil { log.Fatal(err.Error()) }
		rhoVir = RhoC(D, info)
	} else if info.MassDef[end] == 'm' {
		D, err := strconv.ParseFloat(info.MassDef[:end], 64)
		if err != nil { log.Fatal(err.Error()) }
		rhoVir = RhoM(D, info)
	}

	for i := range radius {
		rPhys := math.Pow(mass[i] / (rhoVir * (4*math.Pi/3)), 1/3.0)
		radius[i] = rPhys / info.ScaleFactor
	}

	return radius
}

func ParentIndex(
	x, y, z []float64, radius []float64, info *SimulationInfo,
) (pIdx, upIdx []int) {
	coord := make([][3]float64, len(x))
	for i := range coord { coord[i] = [3]float64{x[i], y[i], z[i]} }

	pIdx, upIdx = make([]int, len(x)), make([]int, len(x))
	for i := range pIdx { pIdx[i], upIdx[i] = -1, -1 }

	finder := box.NewFinder(info.L, coord)

	for i := range coord {
		children := finder.Find(coord[i], radius[i])
		for _, j := range children {
			if radius[j] >= radius[i] { continue }
			if pIdx[j] == -1 || radius[i] < radius[pIdx[j]] { pIdx[j] = i }
			if upIdx[j] == -1 || radius[i] > radius[upIdx[j]] { upIdx[j] = i }
		}
	}

	return pIdx, upIdx
}

func main() {
	// Read command line arguments
	if len(os.Args) != 3 {
		log.Fatalf(`Correct usage:
    ./find_subhaloes path/to/config path/to/catalogue > path/to/output`)
	}

	configFname, catalogueFname := os.Args[1], os.Args[2]
	info := ReadSimulationInfo(configFname)
	reader := catalogue.TextFile(catalogueFname)

	// Read halo catalogue
	icols := reader.ReadInts([]int{int(info.IDIndex)})
	fcols := reader.ReadFloat64s([]int{
		int(info.MassIndex), int(info.PositionIndex[0]),
		int(info.PositionIndex[1]), int(info.PositionIndex[2]),
	})
	id, mass, x, y, z := icols[0], fcols[0], fcols[1], fcols[2], fcols[3]

	// Find halo radii
	radius := Radius(mass, info)
	for i := range radius { radius[i] *= info.RadiusMult }

	// Find parent indices
	pIdx, upIdx := ParentIndex(x, y, z, radius, info)
	pid, upid := make([]int, len(pIdx)), make([]int, len(upIdx))
	for i := range id {
		if pIdx[i] == -1 { 
			pid[i], upid[i] = -1, -1
		} else {
			pid[i], upid[i] = id[pIdx[i]], id[upIdx[i]]
		}
	}

	// Print catalogue
	radiusName := fmt.Sprintf("R_%s", info.MassDef)
	if info.RadiusMult != 1.0 {
		radiusName = fmt.Sprintf("%f * %s", info.RadiusMult, radiusName)
	}

	fmt.Printf("# Subhalo classifications within %s\n", radiusName)
	fmt.Printf("# 0 - ID\n")
	fmt.Printf("# 1 - Parent Index\n")
	fmt.Printf("# 2 - UParent Index\n")
	fmt.Printf("# 3 - PID\n")
	fmt.Printf("# 4 - UPID\n")
	for i := range id {
		fmt.Printf("%10d %10d %10d %10d %10d\n",
			id[i], pIdx[i], upIdx[i], pid[i], upid[i])
	}
}
