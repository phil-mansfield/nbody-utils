package catalogue

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"unsafe"
)

type binhReader struct {
	hd *BinhHeader
	blockOffsets []int
	enc *BinhEncoder
	rd io.ReadSeeker
}

func newBinhReader(fname string) *binhReader {
	rd := &binhReader{ }

	var err error
	rd.rd, err = os.Open(fname)
	if err != nil { panic(err.Error()) }

	rd.hd = readBinhHeader(rd.rd)
	headerOffsets := blockHeaderOffsets(rd.rd, rd.hd)
	rd.enc = &BinhEncoder{ }
	
	rd.hd.BlockHaloes = make([]int64, rd.Blocks())
	rd.hd.BlockFlags = make([][]ColumnFlag, rd.Blocks())
	rd.hd.BlockKeys = make([][]int64, rd.Blocks())
	rd.blockOffsets = make([]int, rd.Blocks())

	for i := 0; i < rd.Blocks(); i++ {
		rd.rd.Seek(int64(headerOffsets[i]), 0)
		binary.Read(rd.rd, binary.LittleEndian, &rd.hd.BlockHaloes[i])
		rd.hd.Haloes += rd.hd.BlockHaloes[i]
		rd.hd.BlockFlags[i] = make([]ColumnFlag, rd.hd.Columns)
		binary.Read(rd.rd, binary.LittleEndian, rd.hd.BlockFlags[i])
		rd.hd.BlockKeys[i] = make([]int64, rd.hd.Columns)
		binary.Read(rd.rd, binary.LittleEndian, rd.hd.BlockKeys[i])

		rd.blockOffsets[i] = headerOffsets[i] + 8 * int(2*rd.hd.Columns + 1)
	}

	return rd
}

func readBinhHeader(rd io.ReadSeeker) *BinhHeader {
	hd := &BinhHeader{ }
	
	binary.Read(rd, binary.LittleEndian, &hd.BinhFixedWidthHeader)
	
	if hd.Version != BinhVersion {
		panic(fmt.Sprintf("BinhVersion = %d, but reader is version %d.",
			hd.Version, BinhVersion))
	}

	hd.Deltas = make([]float64, hd.Columns)
	hd.ColumnSkipped = make([]uint8, hd.Columns)
	hd.TextHeader = make([]byte, hd.TextHeaderLength)
	hd.TextColumnNames = make([]byte, hd.TextColumnNamesLength)

	binary.Read(rd, binary.LittleEndian, hd.Deltas)
	binary.Read(rd, binary.LittleEndian, hd.ColumnSkipped)
	binary.Read(rd, binary.LittleEndian, hd.TextHeader)
	binary.Read(rd, binary.LittleEndian, hd.TextColumnNames)

	hd.ColumnLookup = map[string]int{ }
	names := strings.Split(string(hd.TextColumnNames), ",")
	for i := range names { hd.ColumnLookup[names[i]] = i }
	return hd
}

func blockHeaderOffsets(rd io.ReadSeeker, hd *BinhHeader) []int {
	headerSize := int(unsafe.Sizeof(BinhFixedWidthHeader{})) +
		len(hd.Deltas) * 8 +
		len(hd.ColumnSkipped) +
		len(hd.TextHeader) +
		len(hd.TextColumnNames)

	headerOffsets := make([]int, hd.Blocks)
	headerOffsets[0] = headerSize

	haloes := int64(0)
	flags := make([]ColumnFlag, hd.Columns)

	for i := 1; i < len(headerOffsets); i++ {
		rd.Seek(int64(headerOffsets[i - 1]), 0)
		binary.Read(rd, binary.LittleEndian, &haloes)
		binary.Read(rd, binary.LittleEndian, &flags)

		flagSize := 0
		for j := range flags {
			if hd.ColumnSkipped[j] == 0 { flagSize += flags[j].Size() }
		}
		size := 8 + 8*int(hd.Columns) + 8*int(hd.Columns) +
			flagSize * int(haloes)
		headerOffsets[i] = headerOffsets[i - 1] + size
	}

	return headerOffsets
}

func (rd *binhReader) ReadInts(
	columns interface{}, optBuf ...[][]int,
) [][]int {
	cols := rd.columnIndices(columns)
	bufs := cleanIntBuffer(optBuf, len(cols), int(rd.hd.Haloes))

	start, end := 0, 0
	for block := 0; block < int(rd.hd.Blocks); block++ {
		start = end
		end += int(rd.hd.BlockHaloes[block])
		for i := range cols {
			col := cols[i]
			rd.readIntColumn(block, col, bufs[i][start:end])
		}
	}

	return bufs
}

func (rd *binhReader) ReadFloat64s(
	columns interface{}, optBuf ...[][]float64,
) [][]float64 {
	cols := rd.columnIndices(columns)
	bufs := cleanFloat64Buffer(optBuf, len(cols), int(rd.hd.Haloes))

	start, end := 0, 0
	for block := 0; block < int(rd.hd.Blocks); block++ {
		start = end
		end += int(rd.hd.BlockHaloes[block])
		for i := range cols {
			col := cols[i]
			rd.readFloat64Column(block, col, bufs[i][start:end])
		}
	}

	return bufs
}
func (rd *binhReader) ReadFloat32s(
	columns interface{}, optBuf ...[][]float32,
) [][]float32 {
	cols := rd.columnIndices(columns)
	bufs := cleanFloat32Buffer(optBuf, len(cols), int(rd.hd.Haloes))

	start, end := 0, 0
	for block := 0; block < int(rd.hd.Blocks); block++ {
		start = end
		end += int(rd.hd.BlockHaloes[block])
		for i := range cols {
			col := cols[i]
			rd.readFloat32Column(block, col, bufs[i][start:end])
		}
	}

	return bufs
}
	
func (rd *binhReader) Blocks() int {
	return int(rd.hd.Blocks)
}
	
func (rd *binhReader) ReadIntBlock(
	columns interface{}, block int, optBuf ...[][]int,
) [][]int {
	cols := rd.columnIndices(columns)
	bufs := cleanIntBuffer(optBuf, len(cols), int(rd.hd.BlockHaloes[block]))
	for i := range cols {
		rd.readIntColumn(block, cols[i], bufs[i])
	}

	return bufs
}

func (rd *binhReader)  ReadFloat64Block(
	columns interface{}, block int, optBuf ...[][]float64,
) [][]float64 {
	cols := rd.columnIndices(columns)
	bufs := cleanFloat64Buffer(optBuf, len(cols), int(rd.hd.BlockHaloes[block]))
	for i := range cols {
		rd.readFloat64Column(block, cols[i], bufs[i])
	}

	return bufs
}

func (rd *binhReader) ReadFloat32Block(
	columns interface{}, block int, optBuf ...[][]float32,
) [][]float32 {
	cols := rd.columnIndices(columns)
	bufs := cleanFloat32Buffer(optBuf, len(cols), int(rd.hd.BlockHaloes[block]))
	for i := range cols {
		rd.readFloat32Column(block, cols[i], bufs[i])
	}

	return bufs
}

func (rd *binhReader) columnIndices(columns interface{}) []int {
	if intCols, ok := columns.([]int); ok {
		return intCols
	} else if strCols, ok := columns.([]string); ok {
		idxs := make([]int, len(strCols))
		for i := range strCols {
			names := strings.ToLower(strings.Trim(strCols[i], " "))
			if idx, ok := rd.hd.ColumnLookup[names]; ok {
				idxs[i] = idx
			} else {
				panic(fmt.Sprintf("Name '%s' not in columns.", strCols[i]))
			}
		}
		return idxs
	} 
	panic("Columns argument must be []int or []string.")
}

func (rd *binhReader) columnByteIndex(block, col int) int {
	if rd.hd.ColumnSkipped[col] == 1 {
		panic(fmt.Sprintf("Column %d was skipped for this binh file.", col))
	}

	blockOffset := rd.blockOffsets[block]

	startTypeSize := 0
	for i := 0; i < col; i++ {
		if rd.hd.ColumnSkipped[i] == 0 {
			startTypeSize += rd.hd.BlockFlags[block][i].Size()
		}
	}

	return int(rd.hd.BlockHaloes[block])*startTypeSize + int(blockOffset)
}

func cleanIntBuffer(bufArg [][][]int, cols, n int) [][]int {
	var bufs [][]int
	if len(bufArg) == 0 {
		bufs = make([][]int, cols)
	} else {
		bufs = bufArg[0]
	}

	for i := range bufs {
		bufs[i] = bufs[i][:cap(bufs[i])]
		needed := n - len(bufs[i])
		if needed > 0 {
			bufs[i] = append(bufs[i], make([]int, needed)...)
		} else {
			bufs[i] = bufs[i][:n]
		}
	}

	return bufs
}

func cleanFloat64Buffer(bufArg [][][]float64, cols, n int) [][]float64 {
	var bufs [][]float64
	if len(bufArg) == 0 {
		bufs = make([][]float64, cols)
	} else {
		bufs = bufArg[0]
	}

	for i := range bufs {
		bufs[i] = bufs[i][:cap(bufs[i])]
		needed := n - len(bufs[i])
		if needed > 0 {
			bufs[i] = append(bufs[i], make([]float64, needed)...)
		} else {
			bufs[i] = bufs[i][:n]
		}
	}

	return bufs
}


func cleanFloat32Buffer(bufArg [][][]float32, cols, n int) [][]float32 {
	var bufs [][]float32
	if len(bufArg) == 0 {
		bufs = make([][]float32, cols)
	} else {
		bufs = bufArg[0]
	}

	for i := range bufs {
		bufs[i] = bufs[i][:cap(bufs[i])]
		needed := n - len(bufs[i])
		if needed > 0 {
			bufs[i] = append(bufs[i], make([]float32, needed)...)
		} else {
			bufs[i] = bufs[i][:n]
		}
	}

	return bufs
}


func (rd *binhReader) readIntColumn(block, col int, out []int) {
	start := rd.columnByteIndex(block, col)
	flag, key := rd.hd.BlockFlags[block][col], rd.hd.BlockKeys[block][col]
	rd.rd.Seek(int64(start), 0)
	rd.enc.DecodeInts(flag, key, rd.rd, out)
}

func (rd *binhReader) readFloat64Column(block, col int, out []float64) {
	start := rd.columnByteIndex(block, col)
	flag, key := rd.hd.BlockFlags[block][col], rd.hd.BlockKeys[block][col]
	delta := rd.hd.Deltas[col]
	rd.rd.Seek(int64(start), 0)
	rd.enc.DecodeFloat64s(flag, delta, key, rd.rd, out)
}

func (rd *binhReader) readFloat32Column(block, col int, out []float32) {
	start := rd.columnByteIndex(block, col)
	flag, key := rd.hd.BlockFlags[block][col], rd.hd.BlockKeys[block][col]
	delta := float32(rd.hd.Deltas[col])
	rd.rd.Seek(int64(start), 0)
	rd.enc.DecodeFloat32s(flag, delta, key, rd.rd, out)
}
