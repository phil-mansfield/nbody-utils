package io

import (
	"encoding/binary"
	"fmt"
	"strings"
	"os"

	"unsafe"
)

var (
	BinHOrder = binary.LittleEndian
)

type BinHFile struct {
	fname string
	hd BinHHeader
	intColumns []int64
	header string
	names []string

	fNameIndex map[string]int
	iNameIndex map[string]int
}

type BinHHeader struct {
	Mode int64
	Haloes, Columns, MassColumn int64
	IntColumns, NamesLength, TextHeaderLength int64	
}

// OpenBinH
func OpenBinH(name string) *BinHFile {
	bf := &BinHFile{}
	bf.fname = name

	f, err := os.Open(name)
	if err != nil { panic(err.Error()) }
	defer f.Close()

	hd := &BinHHeader{}
	binary.Read(f, BinHOrder, hd)

	bf.intColumns = make([]int64, hd.IntColumns)
	textNames := make([]byte, hd.NamesLength)
	textHeader := make([]byte, hd.TextHeaderLength)

	binary.Read(f, BinHOrder, bf.intColumns)
	binary.Read(f, BinHOrder, textNames)
	binary.Read(f, BinHOrder, textHeader)

	bf.hd = *hd
	bf.names = strings.Split(string(textNames), "\n")
	bf.header = string(textHeader)


	bf.iNameIndex, bf.fNameIndex = map[string]int{}, map[string]int{}
	for i := int64(0); i < bf.hd.Columns; i++ {
		float := true
		for j := range bf.intColumns {
			if i == bf.intColumns[j] {
				float = false
				break
			}
		}

		if float {
			bf.fNameIndex[bf.names[i]] = int(i)
		} else {
			bf.iNameIndex[bf.names[i]] = int(i)
		}
	}

	return bf
}

func (bf *BinHFile) ReadHeader() string { return bf.header }

func (bf *BinHFile) ReadNames() []string { return bf.names }

func (bf *BinHFile) floatOffset(col int) int64 {
	idx := int64(col)
	for i := range bf.intColumns {
		if bf.intColumns[i] == int64(col) {
			panic(fmt.Sprintf(
				"Tried to read column %d as a float, but it's an int.", col,
			))
		} else if bf.intColumns[i] < int64(col) {
			idx--
		}
	}

	offset := int64(unsafe.Sizeof(BinHHeader{})) + 
		8*bf.hd.IntColumns +
		bf.hd.NamesLength +
		bf.hd.TextHeaderLength +
		8*bf.hd.Haloes*bf.hd.IntColumns +
		4*int64(idx)*bf.hd.Haloes

	return offset
}

func (bf *BinHFile) intOffset(col int) int64 {
	idx := int64(-1)
	for i := range bf.intColumns {
		if bf.intColumns[i] == int64(col) { idx = int64(i) }
	}
	if idx == -1 {
		panic(fmt.Sprintf("Column %d is not an integer column.", col))
	}

	offset := int64(unsafe.Sizeof(BinHHeader{})) + 
		8*bf.hd.IntColumns +
		bf.hd.NamesLength +
		bf.hd.TextHeaderLength +
		8*int64(idx)*bf.hd.Haloes

	return offset
}

// ReadFloats
func (bf *BinHFile) ReadFloats(cols []int) [][]float64 {
	offsets := make([]int64, len(cols))
	for i := range cols { offsets[i] = bf.floatOffset(cols[i]) }

	out32 := make([][]float32, len(cols))
	for i := range out32 { out32[i] = make([]float32, bf.hd.Haloes) }

	f, err := os.Open(bf.fname)
	if err != nil { panic(err.Error()) }
	defer f.Close()

	for i := range out32 {
		_, err = f.Seek(offsets[i], 0)
		if err != nil { panic(err.Error()) }
		err = binary.Read(f, BinHOrder, out32[i])
		if err != nil { panic(err.Error()) }
	}

	out64 := make([][]float64, len(cols))
	for i := range out64 {
		out64[i] = make([]float64, len(out32[i]))
		for j := range out64[i] {
			out64[i][j] = float64(out32[i][j])
		}
	}

	return out64
}

func (bf *BinHFile) ReadInts(cols []int) [][]int {
	offsets := make([]int64, len(cols))
	for i := range cols { offsets[i] = bf.intOffset(cols[i]) }

	out64 := make([][]int64, len(cols))
	for i := range out64 { out64[i] = make([]int64, bf.hd.Haloes) }

	f, err := os.Open(bf.fname)
	if err != nil { panic(err.Error()) }
	defer f.Close()

	for i := range out64 {
		_, err = f.Seek(offsets[i], 0)
		if err != nil { panic(err.Error()) }
		err = binary.Read(f, BinHOrder, out64[i])
		if err != nil { panic(err.Error()) }
	}

	outInt := make([][]int, len(cols))
	for i := range outInt {
		outInt[i] = make([]int, len(out64[i]))
		for j := range outInt[i] {
			outInt[i][j] = int(out64[i][j])
		}
	}

	return outInt
}

func (bf *BinHFile) ReadIntsByName(names []string) [][]int {
	cols := []int{}
	for i := range names {
		col, iOk := bf.iNameIndex[names[i]]
		_, fOk := bf.fNameIndex[names[i]]
		if fOk {
			panic(fmt.Sprintf(
				"'%s' is a float column, not an int column", names[i],
			))
		} else if !iOk {
			panic(fmt.Sprintf("'%s' isn't a valid column name", names[i]))
		}
		cols = append(cols, col)
	}

	return bf.ReadInts(cols)
}

func (bf *BinHFile) ReadFloatsByName(names []string) [][]float64 {
	cols := []int{}
	for i := range names {
		_, iOk := bf.iNameIndex[names[i]]
		col, fOk := bf.fNameIndex[names[i]]
		if iOk {
			panic(fmt.Sprintf(
				"'%s' is an int column, not a float column", names[i],
			))
		} else if !fOk {
			panic(fmt.Sprintf("'%s' isn't a valid column name", names[i]))
		}
		cols = append(cols, col)
	}

	return bf.ReadFloats(cols)
}
