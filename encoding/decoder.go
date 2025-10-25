package encoding

import (
	"bytes"
	"encoding/binary"
	"math"
)

type XORDecoder struct {
	reader        *bytes.Reader
	first         bool
	prevValue     float64
	prevValueBits uint64
}

func NewXORDecoder(data []byte) *XORDecoder {
	return &XORDecoder{
		reader: bytes.NewReader(data),
		first:  true,
	}
}

func (d *XORDecoder) Decode() (float64, error) {
	if d.first {
		var valueBits uint64
		err := binary.Read(d.reader, binary.LittleEndian, &valueBits)
		if err != nil {
			return 0, err
		}

		value := math.Float64frombits(valueBits)
		d.prevValue = value
		d.prevValueBits = valueBits
		d.first = false
		return value, nil
	}

	delta, err := d.readVarUint()
	if err != nil {
		return 0, err
	}

	currentBits := d.prevValueBits ^ delta
	value := math.Float64frombits(currentBits)

	d.prevValue = value
	d.prevValueBits = currentBits

	return value, nil
}

func (d *XORDecoder) readVarUint() (uint64, error) {
	var v uint64
	var shift uint

	for {
		b, err := d.reader.ReadByte()
		if err != nil {
			return 0, err
		}

		v |= uint64(b&0x7F) << shift
		shift += 7

		if b&0x80 == 0 {
			break
		}
	}

	return v, nil
}

func (d *XORDecoder) Reset(data []byte) {
	d.reader = bytes.NewReader(data)
	d.first = true
	d.prevValue = 0
	d.prevValueBits = 0
}
