package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/phil-mansfield/shellfish/cmd/catalog"
	ar "github.com/phil-mansfield/nbody-utils/array"
	"github.com/phil-mansfield/nbody-utils/config"
	nbio "github.com/phil-mansfield/nbody-utils/io"
)

type Config struct {
	ParticleMass float64
	MinParticles int64
	Columns int64
	HeaderLines int64
	MassColumn int64
	IntColumns []int64
	ColumnNames []string
}

// PrintError prints out an error message and example config file.
func PrintError() {
	fmt.Println(
		`Incorrect arguments. Correct usage:
    convert_catalogue [config_filename] [catalogue_name] [output_name]

    Example config:
[convert_catalogue]
ParticleMass = 1.3e8
MinParticles = 200
Columns = 60
HeaderLines = 0
MassColumn = 10
IntColumns = 0, 10, 15, 20
# ColumnNames = ..., optional`,
	)
}

// ParseConfig parses a config file and does basic error analysis.
func ParseConfig(fname string) Config {
	c := Config{}
	vars := config.NewConfigVars("convert_catalogue")
	vars.Float(&c.ParticleMass, "ParticleMass", -1.0)
	vars.Int(&c.MinParticles, "MinParticles", 200)
	vars.Int(&c.Columns, "Columns", -1)
	vars.Int(&c.HeaderLines, "HeaderLines", 0)
	vars.Int(&c.MassColumn, "MassColumn", -1)
	vars.Ints(&c.IntColumns, "IntColumns", []int64{})
	vars.Strings(&c.ColumnNames, "ColumnNames", []string{})

	err := config.ReadConfig(fname, vars)
	if err != nil { 
		panic(err.Error())
	} else if c.ParticleMass < 0 {
		panic("ParticleMass not given.")
	} else if c.Columns < 0 {
		panic("Columns not given.")
	} else if c.MassColumn < 0 {
		panic("MassColumn not given.")
	}

	if len(c.ColumnNames) == 0 {
		c.ColumnNames = make([]string, c.Columns)
		for i := range c.ColumnNames {
			c.ColumnNames[i] = fmt.Sprintf("%d", i)
		}
	}

	if len(c.ColumnNames) != int(c.Columns) {
		panic(fmt.Sprintf(
			"Columns = %d, but len(ColumnNames) = %d.",
			c.Columns, c.ColumnNames,
		))
	}

	for i := range c.ColumnNames {
		c.ColumnNames[i] = strings.Trim(c.ColumnNames[i], " ")
	}

	max := int64(-1)
	for i := range c.IntColumns {
		if c.IntColumns[i] < max {
			panic("IntColumns isn't sorted.")
		} else if c.IntColumns[i] == max {
			panic("Duplicate entries in IntColumns.")
		} else {
			max = c.IntColumns[i]
		}
	}
	
	return c
}

// SliceNTokens slices b to contain all the text up to and including the nth
// instance of tok along with n - m, where m is the number of instance of tok
// in the returned slice.
func SliceNTokens(b []byte, tok byte, n int) (s []byte, nLeft int) {
	for i := range b {
		if b[i] == tok {
			n--
			if n == 0 { return b[:i+1], 0 }
		}
	}

	return b, n
}


// ReadNTokens reads all the text up to and including the nth instance of tok.
func ReadNTokens(rd io.Reader, tok byte, n int) []byte {
	buf := make([]byte, 1<<10)
	out, slice := []byte{}, []byte{}
	
	for n > 0 {
		nRead, err := io.ReadFull(rd, buf)
		if err == io.EOF { break }
		slice, n = SliceNTokens(buf[:nRead], tok, n)

		out = append(out, slice...)
	}

	return out
}

// ReadTextHeader reads c.HeaderLines from haloFile withou needing to read the
// whole file.
func ReadTextHeader(haloFile string, c Config) []byte {
	f, err := os.Open(haloFile)
	defer f.Close()
	if err != nil { panic(err.Error()) }	
	return ReadNTokens(f, byte('\n'), int(c.HeaderLines))
}

// FormatNames converts the column names into the format they'll be saved to
// disk with.
func FormatNames(c Config) []byte {
	return []byte(strings.Join(c.ColumnNames, "\n"))
}

// CountHaloes counts the number of haloes in the given columns.
func CountHaloes(icols [][]int, fcols [][]float64) int {
	if len(icols) > 0 {
		return len(icols[0])
	} else if len(fcols) > 0 {
		return len(fcols[0])
	}
	panic("No columns in config file.")
}

// ColumnIndices uses Config to split column indices into integer columns and
// float columns.
func ColumnIndices(c Config) (icolIdxs, fcolIdxs []int) {
	isInt := func(x int64) bool {
		for _, col := range c.IntColumns {
			if col == x { return true }
		}
		return false
	}

	for i := int64(0); i < c.Columns; i++ {
		if isInt(i) {
			icolIdxs = append(icolIdxs, int(i))
		} else {
			fcolIdxs = append(fcolIdxs, int(i))
		}
	}

	return icolIdxs, fcolIdxs
}

// ReadColumns reads a text catalog based on the information in Config.
func ReadColumns(haloName string, c Config) (icols [][]int, fcols [][]float64) {
	icolIdxs, fcolIdxs := ColumnIndices(c)
	icols, fcols, err := catalog.ReadFile(haloName, icolIdxs, fcolIdxs)
	if err != nil { panic(err.Error()) }
	return icols, fcols
}

// MakeBinaryHeader initializes a binary header.
func MakeBinaryHeader(
	c Config, nHalo int, textHeader, textNames []byte,
) nbio.BinHHeader {
	return nbio.BinHHeader{
		Haloes: int64(nHalo),
		Columns: c.Columns,
		MassColumn: c.MassColumn,
		IntColumns: int64(len(c.IntColumns)), 
		NamesLength: int64(len(textNames)),
		TextHeaderLength: int64(len(textHeader)),
	}
}

// FColMassIndex returns the index of the mass column within the float columns.
func FColMassIndex(intCols []int64, massCol int64) int {
	fcolMassIndex := massCol
	for i := range intCols {
		if intCols[i] == massCol {
			panic("MassColumn is an IntColumn")
		} else if intCols[i] < massCol {
			fcolMassIndex--
		}
	}
	return int(fcolMassIndex)
}

// CutAndSortColumns applies a mass cut to the haloes and then sorts according
// to the mass column.
func CutAndSortColumns(c Config, icols[][]int, fcols [][]float64) {
	massCol := fcols[FColMassIndex(c.IntColumns, c.MassColumn)]
	cut := ar.Greater(massCol, float64(c.MinParticles)*c.ParticleMass)
	order := ar.IntReverse(ar.QuickSortIndex(ar.Cut(massCol, cut)))
	for i := range icols {
		icols[i] = ar.IntOrder(ar.IntCut(icols[i], cut), order)
	}
	for i := range fcols {
		fcols[i] = ar.Order(ar.Cut(fcols[i], cut), order)
	}
}

// Float64ToFloat32 converts float64 slices to float32 slices.
func Float64ToFloat32(fcols [][]float64) [][]float32 {
	out := make([][]float32, len(fcols))
	
	for i := range fcols {
		out[i] = make([]float32, len(fcols[i]))
		for j := range out[i] {
			out[i][j] = float32(fcols[i][j])
		}
	}

	return out 
}

// IntToInt64 converts int slices to int64 slices.
func IntToInt64(icols [][]int) [][]int64 {
	out := make([][]int64, len(icols))

	for i := range icols {
		out[i] = make([]int64, len(icols[i]))
		for j := range out[i] {
			out[i][j] = int64(icols[i][j])
		}
	}
	return out
}

func main() {
	// I/O nonsense.
	if len(os.Args) != 4 {
		PrintError()
		os.Exit(1)
	}
	configFile, haloFile, outFile := os.Args[1], os.Args[2], os.Args[3]

	c := ParseConfig(configFile)

	// Process file.

	textHeader := ReadTextHeader(haloFile, c)
	textNames := FormatNames(c)
	icols, fcols := ReadColumns(haloFile, c)

	CutAndSortColumns(c, icols, fcols)
	nHalo := CountHaloes(icols, fcols)
	hd := MakeBinaryHeader(c, nHalo, textHeader, textNames)
	f32cols := Float64ToFloat32(fcols)
	i64cols := IntToInt64(icols)
	
	// Write data.

	f, err := os.Create(outFile)
	if err != nil { panic(err.Error()) }
	defer f.Close()

	order := binary.LittleEndian
	err = binary.Write(f, nbio.BinHOrder, hd)
	if err != nil { panic(err.Error()) }
	err = binary.Write(f, nbio.BinHOrder, c.IntColumns)
	if err != nil { panic(err.Error()) }
	err = binary.Write(f, nbio.BinHOrder, textNames)
	if err != nil { panic(err.Error()) }
	err = binary.Write(f, nbio.BinHOrder, textHeader)
	if err != nil { panic(err.Error()) }

	for i := range icols {
		err = binary.Write(f, order, i64cols[i])
		if err != nil { panic(err.Error()) }
	}
	for i := range f32cols {
		err = binary.Write(f, order, f32cols[i])
		if err != nil { panic(err.Error()) }
	}
}
