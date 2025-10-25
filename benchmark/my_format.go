package benchmark

//
//import (
//	"time"
//	"tsdb/encoding"
//	"tsdb/types"
//)
//
//func convertToTSDBPoints(points []Point) []types.Point {
//	tsdbPoints := make([]types.Point, len(points))
//	for i, p := range points {
//		tsdbPoints[i] = types.Point{
//			Timestamp: p.Timestamp,
//			Value:     p.Value,
//		}
//	}
//	return tsdbPoints
//}
//
//func BenchmarkMyFormat(dataset *SeriesData, datasetName string) BenchmarkResult {
//	if dataset == nil || len(dataset.Points) == 0 {
//		return BenchmarkResult{}
//	}
//
//	totalPoints := len(dataset.Points)
//	tsdbPoints := convertToTSDBPoints(dataset.Points)
//
//	encodeStart := time.Now()
//	compressedTimestamps, compressedValues, err := encoding.CompressPoints(tsdbPoints)
//	encodeTime := time.Since(encodeStart)
//
//	if err != nil {
//		panic("Our format encoding failed: " + err.Error())
//	}
//
//	totalEncodedSize := len(compressedTimestamps) + len(compressedValues)
//
//	decodeStart := time.Now()
//	decodedPoints, err := encoding.DecompressPoints(compressedTimestamps, compressedValues, totalPoints)
//	decodeTime := time.Since(decodeStart)
//
//	if err != nil {
//		panic("Our format decoding failed: " + err.Error())
//	}
//
//	if len(decodedPoints) != totalPoints {
//		panic("Our format data corruption")
//	}
//
//	encodeSpeed := float64(totalPoints) / encodeTime.Seconds()
//	decodeSpeed := float64(totalPoints) / decodeTime.Seconds()
//
//	return BenchmarkResult{
//		Format:      "OUR_FORMAT",
//		Dataset:     datasetName,
//		DataSize:    totalPoints,
//		EncodedSize: totalEncodedSize,
//		EncodeTime:  encodeTime,
//		DecodeTime:  decodeTime,
//		EncodeSpeed: encodeSpeed,
//		DecodeSpeed: decodeSpeed,
//	}
//}
