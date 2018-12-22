package main

import (
	"fmt"
	"os"

	_ "github.com/phil-mansfield/nbody-utils/array"
	"github.com/phil-mansfield/shellfish/cmd/catalog"
)

const (
	MinParticles = 200
)

func PrintError() {
	fmt.Println(
		`Incorrect arguments. Correct usage:
    convert_catalogue [config_filename] [halo_filename] [output_name]

    The config file should have particle mass on the first line, the number of
    columns on the second, and the index of Mpeak on the third line. All the
    remaining lines are indices of columns with integer data. Use 0-indexing.
`,
	)
}

func ParseConfig(fname string) (mp float64, cols, mpeakCol int, intCols []int) {
	_, fcols, err := catalog.ReadFile(fname)
	vals := fcols[0]

	if len(vals) < 3 {
		panic(fmt.Sprintf("Only %d lines in %s.", len(vals), fname))
	}

	mp := vals[0]
	cols := int(vals[1])
	mpeakCol := int(vals[2])
	if mpeakCol >= cols {
		panic(fmt.Sprintf("mpeakCol = %d, but cols = %d", mpeakCol, cols))
	}
	
	intCols := make([]int, len(vals) - 3)
	for i := range intCols {
		intCols[i] = int(vals[i + 3])
		if intCols[i] >= cols {
			panic(fmt.Sprintf("intCols[%d] = %d, but cols = %d",
				i, intCol[i], cols))
		}
	}

	return mp, cols, mpeakCol, intCols
}

func main() {
	if len(os.Args) != 4 {
		PrintError()
		os.Exit(1)
	}
	configName, haloName, outName := os.Args[1], os.Args[2], os.Args[3]

	mp, cols, mpeakCol, intCols := ParseConfig(configName)

	fmt.Println(mp)
	fmt.Println(cols)
	fmt.Println(mpeakCol)
	fmt.Println(intCols)
}
