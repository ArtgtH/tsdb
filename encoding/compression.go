package encoding

import (
	"bytes"
	"encoding/binary"
	"tsdb/types"
)

func CompressTimestamps(timestamps []int64) []byte {
	if len(timestamps) == 0 {
		return []byte{}
	}

	var buf bytes.Buffer
	prev := timestamps[0]

	binary.Write(&buf, binary.LittleEndian, prev)

	for i := 1; i < len(timestamps); i++ {
		delta := timestamps[i] - prev
		prev = timestamps[i]

		writeZigZagVarInt(&buf, delta)
	}

	return buf.Bytes()
}

func DecompressTimestamps(data []byte, count int) ([]int64, error) {
	if count == 0 {
		return []int64{}, nil
	}

	buf := bytes.NewReader(data)
	timestamps := make([]int64, count)

	err := binary.Read(buf, binary.LittleEndian, &timestamps[0])
	if err != nil {
		return nil, err
	}

	prev := timestamps[0]
	for i := 1; i < count; i++ {
		delta, err := readZigZagVarInt(buf)
		if err != nil {
			return nil, err
		}

		timestamps[i] = prev + delta
		prev = timestamps[i]
	}

	return timestamps, nil
}

func CompressValues(values []float64) []byte {
	if len(values) == 0 {
		return []byte{}
	}

	encoder := NewXOREncoder()

	for _, value := range values {
		encoder.Encode(value)
	}

	return encoder.Bytes()
}

func DecompressValues(data []byte, count int) ([]float64, error) {
	if count == 0 {
		return []float64{}, nil
	}

	values := make([]float64, count)
	decoder := NewXORDecoder(data)

	for i := 0; i < count; i++ {
		value, err := decoder.Decode()
		if err != nil {
			return nil, err
		}
		values[i] = value
	}

	return values, nil
}

func CompressPoints(points []types.Point) ([]byte, []byte, error) {
	timestamps := make([]int64, len(points))
	values := make([]float64, len(points))

	for i, point := range points {
		timestamps[i] = point.Timestamp
		values[i] = point.Value
	}

	compressedTimestamps := CompressTimestamps(timestamps)
	compressedValues := CompressValues(values)

	return compressedTimestamps, compressedValues, nil
}

func DecompressPoints(compressedTimestamps, compressedValues []byte, pointCount int) ([]types.Point, error) {
	timestamps, err := DecompressTimestamps(compressedTimestamps, pointCount)
	if err != nil {
		return nil, err
	}

	values, err := DecompressValues(compressedValues, pointCount)
	if err != nil {
		return nil, err
	}

	points := make([]types.Point, pointCount)
	for i := 0; i < pointCount; i++ {
		points[i] = types.Point{
			Timestamp: timestamps[i],
			Value:     values[i],
		}
	}

	return points, nil
}

func writeZigZagVarInt(buf *bytes.Buffer, v int64) {
	uv := uint64((v << 1) ^ (v >> 63))
	writeVarUint(buf, uv)
}

func readZigZagVarInt(buf *bytes.Reader) (int64, error) {
	uv, err := readVarUint(buf)
	if err != nil {
		return 0, err
	}

	v := int64(uv >> 1)
	if uv&1 != 0 {
		v = ^v
	}
	return v, nil
}

func writeVarUint(buf *bytes.Buffer, v uint64) {
	for v >= 0x80 {
		buf.WriteByte(byte(v) | 0x80)
		v >>= 7
	}
	buf.WriteByte(byte(v))
}

func readVarUint(buf *bytes.Reader) (uint64, error) {
	var v uint64
	var shift uint

	for {
		b, err := buf.ReadByte()
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
