package catalogue

import (
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"encoding/binary"

	"github.com/phil-mansfield/nbody-utils/config"
	ar "github.com/phil-mansfield/nbody-utils/array"
)

var BinhVerbose = false

type BinhConfig struct {
	ParticleMass float64
	MinParticles int64
	Columns int64
	HeaderLines int64
	MassColumn int64
	ColumnInfo []string
	SkipColumns []string
	Sort bool
}

func ParseBinhConfig(fname string) BinhConfig {
	c := BinhConfig{}
	vars := config.NewConfigVars("binh")
	vars.Float(&c.ParticleMass, "ParticleMass", -1.0)
	vars.Int(&c.MinParticles, "MinParticles", 200)
	vars.Int(&c.Columns, "Columns", -1)
	vars.Int(&c.HeaderLines, "HeaderLines", 0)
	vars.Int(&c.MassColumn, "MassColumn", -1)
	vars.Strings(&c.ColumnInfo, "ColumnInfo", []string{})
	vars.Strings(&c.SkipColumns, "SkipColumns", []string{})
	vars.Bool(&c.Sort, "Sort", false)

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

	if len(c.ColumnInfo) != int(c.Columns) {
		panic(fmt.Sprintf(
			"Columns = %d, but len(ColumnInfo) = %d.",
			c.Columns, len(c.ColumnInfo),
		))
	}

	for i := range c.ColumnInfo {
		c.ColumnInfo[i] = strings.Trim(c.ColumnInfo[i], " ")
	}
	
	return c
}

type ColumnFlag int64
const (
	Float64 ColumnFlag = iota
	Float32 
	Int64
	Int32
	Int16
	Int8
	QFloat64
	QFloat32
	QFloat16
	QFloat8
	QLogFloat64
	QLogFloat32
	QLogFloat16
	QLogFloat8	
)

func (flag ColumnFlag) Size() int {
	switch flag {
	case Float64, Int64, QFloat64, QLogFloat64: return 8
	case Float32, Int32, QFloat32, QLogFloat32: return 4
	case Int16, QFloat16, QLogFloat16: return 2
	case Int8, QFloat8, QLogFloat8: return 1
	}
	panic(fmt.Sprintf("ColumnFlag %d not recognized", flag))
}

type BinhEncoder struct {
	int64Buf []int64
	float64Buf []float64
	int32Buf []int32
	float32Buf []float32
	int16Buf []int16
	int8Buf []int8
}

func (enc *BinhEncoder) int64Buffer(n int) []int64 {
	enc.int64Buf = enc.int64Buf[:cap(enc.int64Buf)]

	if len(enc.int64Buf) < n {
		nNew := n - len(enc.int64Buf)
		enc.int64Buf = append(enc.int64Buf, make([]int64, nNew)...)
	}

	return enc.int64Buf[:n]
}

func (enc *BinhEncoder) float64Buffer(n int) []float64 {
	enc.float64Buf = enc.float64Buf[:cap(enc.float64Buf)]

	if len(enc.float64Buf) < n {
		nNew := n - len(enc.float64Buf)
		enc.float64Buf = append(enc.float64Buf, make([]float64, nNew)...)
	}

	return enc.float64Buf[:n]
}

func (enc *BinhEncoder) int32Buffer(n int) []int32 {
	enc.int32Buf = enc.int32Buf[:cap(enc.int32Buf)]

	if len(enc.int32Buf) < n {
		nNew := n - len(enc.int32Buf)
		enc.int32Buf = append(enc.int32Buf, make([]int32, nNew)...)
	}

	return enc.int32Buf[:n]
}

func (enc *BinhEncoder) float32Buffer(n int) []float32 {
	enc.float32Buf = enc.float32Buf[:cap(enc.float32Buf)]

	if len(enc.float32Buf) < n {
		nNew := n - len(enc.float32Buf)
		enc.float32Buf = append(enc.float32Buf, make([]float32, nNew)...)
	}

	return enc.float32Buf[:n]
}

func (enc *BinhEncoder) int16Buffer(n int) []int16 {
	enc.int16Buf = enc.int16Buf[:cap(enc.int16Buf)]

	if len(enc.int16Buf) < n {
		nNew := n - len(enc.int16Buf)
		enc.int16Buf = append(enc.int16Buf, make([]int16, nNew)...)
	}

	return enc.int16Buf[:n]
}

func (enc *BinhEncoder) int8Buffer(n int) []int8 {
	enc.int8Buf = enc.int8Buf[:cap(enc.int8Buf)]

	if len(enc.int8Buf) < n {
		nNew := n - len(enc.int8Buf)
		enc.int8Buf = append(enc.int8Buf, make([]int8, nNew)...)
	}

	return enc.int8Buf[:n]
}

///////////////
// Int sutff //
///////////////

func (enc *BinhEncoder) EncodeInts(
	flag ColumnFlag, x []int, wr io.Writer,
) (key int64) {
	if len(x) == 0 { return 0 }

	min := x[0]
	for i := range x {
		if min > x[i] { min = x[i] }
	}

	switch flag {
	case Int64:
		buf := enc.int64Buffer(len(x))
		for i := range buf { buf[i] = int64(x[i] - min + math.MinInt64) }
		binary.Write(wr, binary.LittleEndian, buf)
	case Int32:
		buf := enc.int32Buffer(len(x))
		for i := range buf { buf[i] = int32(x[i] - min + math.MinInt32) }
		binary.Write(wr, binary.LittleEndian, buf)
	case Int16:
		buf := enc.int16Buffer(len(x))
		for i := range buf { buf[i] = int16(x[i] - min + math.MinInt16) }
		binary.Write(wr, binary.LittleEndian, buf)
	case Int8:
		buf := enc.int8Buffer(len(x))
		for i := range buf { buf[i] = int8(x[i] - min + math.MinInt8) }
		binary.Write(wr, binary.LittleEndian, buf)
	default:
		panic("Cannot encode ints as floats.")
	}

	return int64(min)
}

func (enc *BinhEncoder) DecodeInts(
	flag ColumnFlag, key int64, rd io.Reader, out []int,
) {
	switch flag {
	case Int64:
		buf := enc.int64Buffer(len(out))
		binary.Read(rd, binary.LittleEndian, buf)
		for i := range buf { 
			out[i] = int(int64(buf[i]) + (key - math.MinInt64))
		}
	case Int32:
		buf := enc.int32Buffer(len(out))
		binary.Read(rd, binary.LittleEndian, buf)
		for i := range buf { 
			out[i] = int(int64(buf[i]) + (key - math.MinInt32))
		}
	case Int16:
		buf := enc.int16Buffer(len(out))
		binary.Read(rd, binary.LittleEndian, buf)
		for i := range buf { 
			out[i] = int(int64(buf[i]) + (key - math.MinInt16))
		}
	case Int8:
		buf := enc.int8Buffer(len(out))
		binary.Read(rd, binary.LittleEndian, buf)
		for i := range buf { 
			out[i] = int(int64(buf[i]) + (key - math.MinInt8))
		}
	default:
		panic("Cannot decode ints as floats.")
	}
}

///////////////////
// Float64 stuff //
///////////////////

func (enc *BinhEncoder) EncodeFloat64s(
	flag ColumnFlag, delta float64, x []float64, wr io.Writer,
) (key int64) {
	if len(x) == 0 { return 0 }

	min := x[0]
	for i := range x {
		if min > x[i] { min = x[i] }
	}

	qMin := int64(math.Floor(min / delta))

	switch flag {
	case Float64:
		binary.Write(wr, binary.LittleEndian, x)
	case Float32:
		buf := enc.float32Buffer(len(x))
		for i := range buf { buf[i] = float32(x[i]) }
		binary.Write(wr, binary.LittleEndian, buf)
	case QFloat64:
		buf := enc.int64Buffer(len(x))
		for i := range buf {
			buf[i] = int64(int64(math.Floor(x[i] / delta)) -
				qMin + math.MinInt64)
		}
		binary.Write(wr, binary.LittleEndian, buf)
	case QFloat32:
		buf := enc.int32Buffer(len(x))
		for i := range buf {
			buf[i] = int32(int64(math.Floor(x[i] / delta)) -
				qMin + math.MinInt32)
		}
		binary.Write(wr, binary.LittleEndian, buf)
	case QFloat16:
		buf := enc.int16Buffer(len(x))
		for i := range buf {
			buf[i] = int16(int64(math.Floor(x[i] / delta)) -
				qMin + math.MinInt16)
		}
		binary.Write(wr, binary.LittleEndian, buf)
	case QFloat8:
		buf := enc.int8Buffer(len(x))
		for i := range buf {
			buf[i] = int8(int64(math.Floor(x[i] / delta)) - qMin + math.MinInt8)
		}
		binary.Write(wr, binary.LittleEndian, buf)
	case QLogFloat64:
		buf := enc.float64Buffer(len(x))
		for i := range buf  { buf[i] = math.Log10(x[i]) }
		return enc.EncodeFloat64s(QFloat64, delta, buf, wr)
	case QLogFloat32:
		buf := enc.float64Buffer(len(x))
		for i := range buf  { buf[i] = math.Log10(x[i]) }
		return enc.EncodeFloat64s(QFloat32, delta, buf, wr)
	case QLogFloat16:
		buf := enc.float64Buffer(len(x))
		for i := range buf  { buf[i] = math.Log10(x[i]) }
		return enc.EncodeFloat64s(QFloat16, delta, buf, wr)
	case QLogFloat8:
		buf := enc.float64Buffer(len(x))
		for i := range buf  { buf[i] = math.Log10(x[i]) }
		return enc.EncodeFloat64s(QFloat8, delta, buf, wr)
	default:
		panic("Cannot encode floats as ints")
	}

	return qMin
}

func (enc *BinhEncoder) DecodeFloat64s(
	flag ColumnFlag, delta float64, key int64, rd io.Reader, out []float64,
) {
	switch flag {
	case Float64:
		binary.Read(rd, binary.LittleEndian, out)
	case Float32:
		buf := enc.float32Buffer(len(out))
		binary.Read(rd, binary.LittleEndian, buf)
		for i := range buf { out[i] = float64(buf[i]) }
	case QFloat64:
		buf := enc.int64Buffer(len(out))
		binary.Read(rd, binary.LittleEndian, buf)
		for i := range buf {
			out[i] = delta * float64(int64(buf[i]) + (key - math.MinInt64)) +
				rand.Float64() * delta
		}
	case QFloat32:
		buf := enc.int32Buffer(len(out))
		binary.Read(rd, binary.LittleEndian, buf)
		for i := range buf {
			out[i] = delta * float64(int64(buf[i]) + (key - math.MinInt32)) +
				rand.Float64() * delta
		}
	case QFloat16:
		buf := enc.int16Buffer(len(out))
		binary.Read(rd, binary.LittleEndian, buf)
		for i := range buf {
			out[i] = delta * float64(int64(buf[i]) + (key - math.MinInt16)) +
				rand.Float64() * delta
		}
	case QFloat8:
		buf := enc.int8Buffer(len(out))
		binary.Read(rd, binary.LittleEndian, buf)
		for i := range buf {
			out[i] = delta * float64(int64(buf[i]) + (key - math.MinInt8)) +
				rand.Float64() * delta
		}
	case QLogFloat64:
		enc.DecodeFloat64s(QFloat64, delta, key, rd, out)
		for i := range out { out[i] = math.Pow(10, out[i]) }
	case QLogFloat32:
		enc.DecodeFloat64s(QFloat32, delta, key, rd, out)
		for i := range out { out[i] = math.Pow(10, out[i]) }
	case QLogFloat16:
		enc.DecodeFloat64s(QFloat16, delta, key, rd, out)
		for i := range out { out[i] = math.Pow(10, out[i]) }
	case QLogFloat8:
		enc.DecodeFloat64s(QFloat8, delta, key, rd, out)
		for i := range out { out[i] = math.Pow(10, out[i]) }
	default:
		panic("Cannot decode floats as ints")
	}
}

///////////////////
// Float32 stuff //
///////////////////

func (enc *BinhEncoder) EncodeFloat32s(
	flag ColumnFlag, delta float32, x []float32, wr io.Writer,
) (key int64) {
	buf := enc.float64Buffer(len(x))
	for i := range x { buf[i] = float64(x[i]) }
	return enc.EncodeFloat64s(flag, float64(delta), buf, wr)
}

func (enc *BinhEncoder) DecodeFloat32s(
	flag ColumnFlag, delta float32, key int64, rd io.Reader, out []float32,
) {
	buf := enc.float64Buffer(len(out))
	enc.DecodeFloat64s(flag, float64(delta), key, rd, buf)
	for i := range out { out[i] = float32(buf[i]) }
}

////////////////
// Binh stuff //
////////////////

const (
	BinhVersion = 3
	BinhSeed = 1337
)

type BinhFixedWidthHeader struct {
	Version int64 // Version of the binh format being used
	Seed int64 // Seed to use when de-quantizing floats.
	Columns int64 // Number of columns in each blocks
	MassColumn int64 // Column used to order haloes within a blocks
	Blocks int64 // Number of blocks in the file
	TextHeaderLength int64 // Number of bytes in the text header
	TextColumnNamesLength int64 // number of bytes in the 
	MinMass float64 // Approximate smallest mass stored in the file
}

type BinhHeader struct {
	BinhFixedWidthHeader
	// Fields in the actual header
	Deltas []float64
	TextHeader []byte
	TextColumnNames []byte
	ColumnSkipped []uint8
	// Fields stored in the blocks themselves
	Haloes int64
	BlockHaloes []int64
	BlockFlags [][]ColumnFlag
	BlockKeys [][]int64
	// Not stored
	ColumnLookup map[string]int
}

// TextToBinh converts a text file to a binh file.
func TextToBinh(
	inName, outName string, config BinhConfig, textConfig ...TextConfig,
) {
	// Column information
	_, isInt, isLog, deltas := parseColumnInfo(config.ColumnInfo)
	bufIdx, icols, fcols := bufferIndex(isInt)

	// Set up I/O
	wr, err := os.Create(outName)
	if err != nil { panic(err.Error()) }
	checkMem("opening text file")
	rd := TextFile(inName, textConfig...)
	checkMem("created text file reader")

	// Write a blank header for now. We'll come back to this later.
	hd := newBinhHeader(inName, rd.Blocks(), config)
	binary.Write(wr, binary.LittleEndian, hd.BinhFixedWidthHeader)
	binary.Write(wr, binary.LittleEndian, hd.Deltas)
	binary.Write(wr, binary.LittleEndian, hd.ColumnSkipped)
	binary.Write(wr, binary.LittleEndian, hd.TextHeader)
	binary.Write(wr, binary.LittleEndian, hd.TextColumnNames)

	// Set up buffers
	ibuf := make([][]int, len(icols))
	fbuf := make([][]float64, len(fcols))
	colTypes := make([]ColumnFlag, len(isInt))
	colKeys := make([]int64, len(isInt))
	enc := BinhEncoder{ }

	checkMem("finished initialization")

	// Write blocks one by one.
	for block := 0; block < rd.Blocks(); block++ {
		// Read from the text file
		checkMem(fmt.Sprintf("Reading int block %d/%d", block+1, rd.Blocks()))
		ibuf = rd.ReadIntBlock(icols, block, ibuf)
		checkMem(fmt.Sprintf("Reading float block %d/%d", block+1, rd.Blocks()))
		fbuf = rd.ReadFloat64Block(fcols, block, fbuf)
		checkMem(fmt.Sprintf(" %d/%d", block+1, rd.Blocks()))

		// Set up cuts and sorting.
		massCol := fbuf[bufIdx[config.MassColumn]]
		cut := ar.Geq(massCol, hd.MinMass)
		nHaloes := int64(0)
		for i := range cut {
			if cut[i] { nHaloes++ }
		}

		binary.Write(wr, binary.LittleEndian, nHaloes)
			
		runtime.GC()

		// Find column types
		for col := 0; col < len(isInt); col++ {
			// TODO: buffer the cuts here
			runtime.GC()

			if isInt[col] {
				vals := ar.IntCut(ibuf[bufIdx[col]], cut)
				colTypes[col], colKeys[col] = intColumnType(vals)
			} else {
				vals := ar.Cut(fbuf[bufIdx[col]], cut)
				if isLog[col] {
					colTypes[col], colKeys[col] = logFloat64ColumnType(
						vals, deltas[col],
					)
				} else {
					colTypes[col], colKeys[col] = float64ColumnType(
						vals, deltas[col],
					)
				}
			}
		}

		// Write column information
		binary.Write(wr, binary.LittleEndian, colTypes)
		binary.Write(wr, binary.LittleEndian, colKeys)
		 
		// Write columns
		for col := 0; col < len(isInt); col++ {
			// TODO: buffer the cuts here.

			if hd.ColumnSkipped[col] == 1 { continue }
			runtime.GC()

			if isInt[col] {
				vals := ar.IntCut(ibuf[bufIdx[col]], cut)
				enc.EncodeInts(colTypes[col], vals, wr)
			} else {
				vals := ar.Cut(fbuf[bufIdx[col]], cut)
				enc.EncodeFloat64s(
					colTypes[col], deltas[col], vals, wr,
				)
			}
		}
	}
}

func newBinhHeader(inName string, blocks int, config BinhConfig) *BinhHeader {

	names, _, _, deltas := parseColumnInfo(config.ColumnInfo)

	if len(names) != int(config.Columns) {
		panic(fmt.Sprintf("len(ColunmInfo) = %d, but Columns = %d", len(names), config.Columns))
	}

	hd := &BinhHeader{
		Deltas: deltas,
		TextColumnNames: []byte(strings.Join(names, ",")),
		TextHeader: readTextHeader(inName, int(config.HeaderLines)),
		ColumnSkipped: make([]uint8, len(names)),
		ColumnLookup: map[string]int{},
	}

	for i := range names {
		hd.ColumnLookup[names[i]] = i
	}

	for i := range config.SkipColumns {
		name := strings.ToLower(strings.Trim(config.SkipColumns[i], " "))
		hd.ColumnSkipped[hd.ColumnLookup[name]] = 1
	}
	
	hd.BinhFixedWidthHeader = BinhFixedWidthHeader{
		Version: BinhVersion,
		Seed: BinhSeed,
		Columns: int64(len(config.ColumnInfo)),
		MassColumn: config.MassColumn,
		Blocks: int64(blocks),
		TextHeaderLength: int64(len(hd.TextHeader)),
		TextColumnNamesLength: int64(len(hd.TextColumnNames)),
		MinMass: float64(config.MinParticles)*config.ParticleMass,
	}
	
	return hd
}

func readTextHeader(haloFile string, headerLines int) []byte {
	f, err := os.Open(haloFile)
	defer f.Close()
	if err != nil { panic(err.Error()) }	
	return readNTokens(f, byte('\n'), headerLines)
}


func readNTokens(rd io.Reader, tok byte, n int) []byte {
	buf := make([]byte, 1<<10)
	out, slice := []byte{}, []byte{}
	
	for n > 0 {
		nRead, err := io.ReadFull(rd, buf)
		if err == io.EOF { break }
		slice, n = sliceNTokens(buf[:nRead], tok, n)

		out = append(out, slice...)
	}

	return out
}

func sliceNTokens(b []byte, tok byte, n int) (s []byte, nLeft int) {
	for i := range b {
		if b[i] == tok {
			n--
			if n == 0 { return b[:i+1], 0 }
		}
	}

	return b, n
}

func parseColumnInfo(info []string) (
	names []string, isInt, isLog []bool, deltas []float64,
) {
	names = make([]string, len(info))
	isInt = make([]bool, len(info))
	isLog = make([]bool, len(info))
	deltas = make([]float64, len(info))

	for i, column := range info {
		tokens := strings.Split(column, ":")
		switch len(tokens) {
		case 0, 1:
			panic(fmt.Sprintf("column '%s' not given a type.", column))
		case 3:
			var err error
			deltas[i], err = strconv.ParseFloat(strings.Trim(tokens[2]," "), 64)
			if err != nil {
				panic(fmt.Sprintf("column '%s': %s", column, err.Error()))
			}
			fallthrough
		case 2:
			names[i] = strings.ToLower(strings.Trim(tokens[0], " "))
			switch strings.ToLower(strings.Trim(tokens[1], " ")) {
			case "int": isInt[i] = true
			case "float":
			case "log": isLog[i] = true
			default:
				panic(fmt.Sprintf(
					"Unrecognized type '%s' in column '%s'.", tokens[1], column,
				))
			}
		default:
			panic(fmt.Sprintf("column '%s' has too many annotations", column))
		}		
	}

	return names, isInt, isLog, deltas
}

func bufferIndex(isInt []bool) (bufIdx, icols, fcols []int) {
	bufIdx = make([]int, len(isInt))

	iIdx, fIdx := 0, 0
	for i := range isInt {
		if isInt[i] {
			bufIdx[i] = iIdx
			icols = append(icols, i)
			iIdx++
		} else {
			bufIdx[i] = fIdx
			fcols = append(fcols, i)
			fIdx++
		}
	}

	return bufIdx, icols, fcols
}

func rangeToIntType(min, max int) ColumnFlag {
	// Need to do this to avoid overflow
	var (
		low, scaledRange int64
	)
	if min < 0 {
		// It's possible that max - min > math.MaxInt64
		scaledRange = int64(max) - (int64(min) - math.MinInt64)
		low = math.MinInt64
	} else {
		scaledRange = int64(max) - int64(min)
		low = 0
	}

	switch {
	case scaledRange < low + (math.MaxInt8 - math.MinInt8 + 1):
		return Int8
	case scaledRange < low + (math.MaxInt16 - math.MinInt16 + 1):
		return Int16
	case scaledRange < low + (math.MaxInt32 - math.MinInt32 + 1):
		return Int32
	default:
		return Int64
	}
}

func intColumnType(x []int) (flag ColumnFlag, key int64) {
	min, max := x[0], x[0]
	for i := range x {
		if x[i] < min { min = x[i] }
		if x[i] > max { max = x[i] }
	}

	return rangeToIntType(min, max), int64(min)
}

func float64ColumnType(x []float64, delta float64) (
	flag ColumnFlag, key int64,
) {
	if delta == 0 { return Float32, 0 }

	min, max := x[0], x[0]
	for i := range x {
		if x[i] < min { min = x[i] }
		if x[i] > max { max = x[i] }
	}

	qMin, qMax := int(math.Floor(min / delta)), int(math.Floor(max / delta))

	flag = rangeToIntType(qMin, qMax)

	switch flag {
	case Int64: return Float32, 0 // Don't even bother
	case Int32: return Float32, 0 // Don't even bother
	case Int16: return QFloat16, int64(qMin)
	case Int8: return QFloat8, int64(qMin)
	}
	panic("Impossible")
}

func logFloat64ColumnType(x []float64, delta float64) (
	flag ColumnFlag, key int64,
) {
	if delta == 0 { return Float32, 0 }

	min, max := x[0], x[0]
	for i := range x {
		if x[i] <= 0 {
			return Float32, 0
		}
		if x[i] < min { min = x[i] }
		if x[i] > max { max = x[i] }
	}

	qMin := int(math.Floor(math.Log10(min) / delta))
	qMax := int(math.Floor(math.Log10(max) / delta))

	flag = rangeToIntType(qMin, qMax)

	switch flag {
	case Int64: return Float32, 0 // Don't even bother
	case Int32: return Float32, 0 // Don't even bother
	case Int16: return QLogFloat16, int64(qMin)
	case Int8: return QLogFloat8, int64(qMin)
	}
	panic("Impossible")
}

func checkMem(s string) {
	if !BinhVerbose { return }
	log.Println(s)
	ms := runtime.MemStats{ }
	runtime.ReadMemStats(&ms)
	fmt.Printf("Alloc: %5d MB Sys: %5d MB TotalAlloc: %5d MB\n",
		ms.Alloc >> 20, ms.Sys >> 20, ms.TotalAlloc >> 20)
	runtime.GC()
	runtime.ReadMemStats(&ms)
	fmt.Printf("Alloc: %5d MB Sys: %5d MB TotalAlloc: %5d MB\n",
		ms.Alloc >> 20, ms.Sys >> 20, ms.TotalAlloc >> 20)
}
