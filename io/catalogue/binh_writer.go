package catalogue

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"strings"
	"encoding/binary"

	"github.com/phil-mansfield/nbody-utils/config"
)

type BinhConfig struct {
	ParticleMass float64
	MinParticles int64
	Columns int64
	HeaderLines int64
	MassColumn int64
	ColumnInfo []string
	SkipColumns []string
}

func ParseBinhConfig(fname string) BinhConfig {
	c := BinhConfig{}
	vars := config.NewConfigVars("convert_catalogue")
	vars.Float(&c.ParticleMass, "ParticleMass", -1.0)
	vars.Int(&c.MinParticles, "MinParticles", 200)
	vars.Int(&c.Columns, "Columns", -1)
	vars.Int(&c.HeaderLines, "HeaderLines", 0)
	vars.Int(&c.MassColumn, "MassColumn", -1)
	vars.Strings(&c.ColumnInfo, "ColumnInfo", []string{})

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

type ColumnFlag int
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

	qMin := int64(min / delta)

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
			buf[i] = int64(int64(x[i] / delta) - qMin + math.MinInt64)
		}
		binary.Write(wr, binary.LittleEndian, buf)
	case QFloat32:
		buf := enc.int32Buffer(len(x))
		for i := range buf {
			buf[i] = int32(int64(x[i] / delta) - qMin + math.MinInt32)
		}
		binary.Write(wr, binary.LittleEndian, buf)
	case QFloat16:
		buf := enc.int16Buffer(len(x))
		for i := range buf {
			buf[i] = int16(int64(x[i] / delta) - qMin + math.MinInt16)
		}
		binary.Write(wr, binary.LittleEndian, buf)
	case QFloat8:
		buf := enc.int8Buffer(len(x))
		for i := range buf {
			buf[i] = int8(int64(x[i] / delta) - qMin + math.MinInt8)
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

