package encoding

import (
	"bytes"
	"encoding/binary"
	"math"
)

type XOREncoder struct {
	prevValue     float64
	prevValueBits uint64
	first         bool
	buffer        *bytes.Buffer
}

func NewXOREncoder() *XOREncoder {
	return &XOREncoder{
		first:  true,
		buffer: &bytes.Buffer{},
	}
}

func (e *XOREncoder) Encode(value float64) {
	valueBits := math.Float64bits(value)

	if e.first {
		binary.Write(e.buffer, binary.LittleEndian, valueBits)
		e.prevValue = value
		e.prevValueBits = valueBits
		e.first = false
		return
	}

	delta := valueBits ^ e.prevValueBits

	e.writeVarUint(delta)

	e.prevValue = value
	e.prevValueBits = valueBits
}

func (e *XOREncoder) Bytes() []byte {
	return e.buffer.Bytes()
}

func (e *XOREncoder) Reset() {
	e.first = true
	e.prevValue = 0
	e.prevValueBits = 0
	e.buffer.Reset()
}

func (e *XOREncoder) writeVarUint(v uint64) {
	for v >= 0x80 {
		e.buffer.WriteByte(byte(v) | 0x80)
		v >>= 7
	}
	e.buffer.WriteByte(byte(v))
}
