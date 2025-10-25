package benchmark

//
//import (
//	"encoding/json"
//	"fmt"
//	"math"
//	"math/rand"
//	"time"
//)
//
//type Point struct {
//	Timestamp int64   `json:"timestamp"`
//	Value     float64 `json:"value"`
//}
//
//type SeriesData struct {
//	Metric string            `json:"metric"`
//	Tags   map[string]string `json:"tags"`
//	Points []Point           `json:"points"`
//}
//
//type BenchmarkResult struct {
//	Format      string
//	Dataset     string
//	DataSize    int
//	EncodedSize int
//	EncodeTime  time.Duration
//	DecodeTime  time.Duration
//	EncodeSpeed float64
//	DecodeSpeed float64
//}
//
//func GenerateDataset(name string, targetPoints int) (*SeriesData, error) {
//	fmt.Printf("Generating %s with %d points...\n", name, targetPoints)
//
//	var series *SeriesData
//
//	switch name {
//	case "metrics_constant":
//		series = generateConstantData(targetPoints, "server_metrics")
//	case "metrics_random":
//		series = generateRandomData(targetPoints, "cpu_usage")
//	case "metrics_sinusoidal":
//		series = generateSinusoidalData(targetPoints, "sine_wave")
//	case "metrics_mixed":
//		series = generateMixedData(targetPoints)
//	default:
//		series = generateMixedData(targetPoints)
//	}
//
//	return series, nil
//}
//
//func generateConstantData(pointsCount int, metric string) *SeriesData {
//	series := &SeriesData{
//		Metric: metric,
//		Tags:   map[string]string{"type": "constant", "source": "sensor_A"},
//		Points: make([]Point, pointsCount),
//	}
//
//	baseTime := time.Now().Unix()
//	for i := 0; i < pointsCount; i++ {
//		series.Points[i] = Point{
//			Timestamp: baseTime + int64(i)*10,
//			Value:     25.5,
//		}
//	}
//
//	return series
//}
//
//func generateRandomData(pointsCount int, metric string) *SeriesData {
//	rand.Seed(42)
//	series := &SeriesData{
//		Metric: metric,
//		Tags:   map[string]string{"type": "random", "host": "server1"},
//		Points: make([]Point, pointsCount),
//	}
//
//	baseTime := time.Now().Unix()
//	for i := 0; i < pointsCount; i++ {
//		series.Points[i] = Point{
//			Timestamp: baseTime + int64(i),
//			Value:     rand.Float64() * 100,
//		}
//	}
//
//	return series
//}
//
//func generateSinusoidalData(pointsCount int, metric string) *SeriesData {
//	series := &SeriesData{
//		Metric: metric,
//		Tags:   map[string]string{"type": "sinusoidal", "freq": "1hz"},
//		Points: make([]Point, pointsCount),
//	}
//
//	baseTime := time.Now().Unix()
//	for i := 0; i < pointsCount; i++ {
//		series.Points[i] = Point{
//			Timestamp: baseTime + int64(i),
//			Value:     math.Sin(float64(i)*0.1)*50 + 50,
//		}
//	}
//
//	return series
//}
//
//func generateMixedData(pointsCount int) *SeriesData {
//	series := &SeriesData{
//		Metric: "mixed_metrics",
//		Tags:   map[string]string{"type": "mixed", "source": "composite"},
//		Points: make([]Point, pointsCount),
//	}
//
//	baseTime := time.Now().Unix()
//	for i := 0; i < pointsCount; i++ {
//		var value float64
//		switch i % 3 {
//		case 0:
//			value = 25.5
//		case 1:
//			value = 50 + 25*math.Sin(float64(i)*0.05)
//		case 2:
//			value = 25 + 50*rand.Float64()
//		}
//
//		series.Points[i] = Point{
//			Timestamp: baseTime + int64(i),
//			Value:     value,
//		}
//	}
//
//	return series
//}
//
//func CalculateRawSize(data *SeriesData) int {
//	jsonData, err := json.Marshal(data)
//	if err != nil {
//		return 0
//	}
//	return len(jsonData)
//}
