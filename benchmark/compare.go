package benchmark

//
//import (
//	"encoding/csv"
//	"fmt"
//	"log"
//	"os"
//	"runtime"
//	"strings"
//	"time"
//)
//
//func StartCompare() {
//	fmt.Println("=== TSDB Format Comparison: Our Format vs Protobuf ===")
//	fmt.Printf("Go version: %s\n", runtime.Version())
//	fmt.Printf("CPU cores: %d\n", runtime.NumCPU())
//	fmt.Println()
//
//	testCases := []struct {
//		name        string
//		pointCount  int
//		description string
//	}{
//		{"small_constant", 100000, "100K points (small)"},
//		{"small_random", 500000, "500K points (small)"},
//		{"medium_constant", 1000000, "1M points (medium)"},
//		{"medium_random", 5000000, "5M points (medium)"},
//		{"large_sinusoidal", 10000000, "10M points (large)"},
//	}
//
//	var results []BenchmarkResult
//
//	fmt.Println("Testing datasets:")
//	for _, tc := range testCases {
//		fmt.Printf("  %s: %s\n", tc.name, tc.description)
//	}
//	fmt.Println()
//
//	for _, tc := range testCases {
//		log.Printf("Testing dataset: %s (%s)", tc.name, tc.description)
//
//		dataset, err := GenerateDataset(tc.name, tc.pointCount)
//		if err != nil {
//			log.Printf("Error generating dataset %s: %v", tc.name, err)
//			continue
//		}
//
//		rawSize := CalculateRawSize(dataset)
//
//		fmt.Printf("  %s: %d points, raw JSON: %.1f MB\n",
//			tc.name, tc.pointCount, float64(rawSize)/1024/1024)
//
//		runtime.GC()
//		time.Sleep(100 * time.Millisecond)
//
//		fmt.Printf("    Testing Our Format... ")
//		ourResult := BenchmarkMyFormat(dataset, tc.name)
//		results = append(results, ourResult)
//		fmt.Printf("✓ (%d bytes, encode: %v, decode: %v)\n",
//			ourResult.EncodedSize, ourResult.EncodeTime, ourResult.DecodeTime)
//
//		dataset = nil
//		runtime.GC()
//		time.Sleep(100 * time.Millisecond)
//
//		dataset, err = GenerateDataset(tc.name, tc.pointCount)
//		if err != nil {
//			log.Printf("Error regenerating dataset %s: %v", tc.name, err)
//			continue
//		}
//
//		fmt.Printf("    Testing Protobuf... ")
//		protoResult := BenchmarkProtobuf(dataset, tc.name)
//		results = append(results, protoResult)
//		fmt.Printf("✓ (%d bytes, encode: %v, decode: %v)\n",
//			protoResult.EncodedSize, protoResult.EncodeTime, protoResult.DecodeTime)
//
//		dataset = nil
//		runtime.GC()
//		time.Sleep(100 * time.Millisecond)
//
//		log.Printf("Completed: %s", tc.name)
//	}
//
//	saveResultsToCSV(results)
//
//	printSummary(results)
//
//	fmt.Println("\nComparison completed! Results saved to benchmark_results.csv")
//}
//
//func saveResultsToCSV(results []BenchmarkResult) {
//	file, err := os.Create("benchmark_results.csv")
//	if err != nil {
//		log.Fatalf("Failed to create CSV file: %v", err)
//	}
//	defer file.Close()
//
//	writer := csv.NewWriter(file)
//	defer writer.Flush()
//
//	writer.Write([]string{
//		"Format",
//		"Dataset",
//		"DataSize_Points",
//		"EncodedSize_Bytes",
//		"EncodeTime_Ns",
//		"DecodeTime_Ns",
//		"EncodeSpeed_PointsPerSec",
//		"DecodeSpeed_PointsPerSec",
//	})
//
//	for _, result := range results {
//		writer.Write([]string{
//			result.Format,
//			result.Dataset,
//			fmt.Sprintf("%d", result.DataSize),
//			fmt.Sprintf("%d", result.EncodedSize),
//			fmt.Sprintf("%d", result.EncodeTime.Nanoseconds()),
//			fmt.Sprintf("%d", result.DecodeTime.Nanoseconds()),
//			fmt.Sprintf("%.2f", result.EncodeSpeed),
//			fmt.Sprintf("%.2f", result.DecodeSpeed),
//		})
//	}
//}
//
//func printSummary(results []BenchmarkResult) {
//	fmt.Println("\n" + strings.Repeat("=", 120))
//	fmt.Println("RESULTS SUMMARY")
//	fmt.Println(strings.Repeat("=", 120))
//	fmt.Printf("%-12s | %-20s | %-10s | %-12s | %-8s | %-8s | %-12s | %-12s\n",
//		"Format", "Dataset", "Points", "Size(bytes)", "Enc(ms)", "Dec(ms)", "Enc(p/s)", "Dec(p/s)")
//	fmt.Println(strings.Repeat("-", 120))
//
//	for _, result := range results {
//		fmt.Printf("%-12s | %-20s | %-10d | %-12d | %-8.1f | %-8.1f | %-12.0f | %-12.0f\n",
//			result.Format,
//			result.Dataset,
//			result.DataSize,
//			result.EncodedSize,
//			float64(result.EncodeTime.Nanoseconds())/1e6,
//			float64(result.DecodeTime.Nanoseconds())/1e6,
//			result.EncodeSpeed,
//			result.DecodeSpeed)
//	}
//
//	fmt.Println("\n" + strings.Repeat("=", 120))
//	fmt.Println("COMPARISON BY DATASET")
//	fmt.Println(strings.Repeat("=", 120))
//
//	datasets := make(map[string][]BenchmarkResult)
//	for _, result := range results {
//		datasets[result.Dataset] = append(datasets[result.Dataset], result)
//	}
//
//	for datasetName, datasetResults := range datasets {
//		if len(datasetResults) != 2 {
//			continue
//		}
//
//		var ourFormat, protobuf BenchmarkResult
//		if datasetResults[0].Format == "OUR_FORMAT" {
//			ourFormat, protobuf = datasetResults[0], datasetResults[1]
//		} else {
//			ourFormat, protobuf = datasetResults[1], datasetResults[0]
//		}
//
//		sizeRatio := float64(ourFormat.EncodedSize) / float64(protobuf.EncodedSize)
//		encodeSpeedRatio := ourFormat.EncodeSpeed / protobuf.EncodeSpeed
//		decodeSpeedRatio := ourFormat.DecodeSpeed / protobuf.DecodeSpeed
//
//		fmt.Printf("\nDataset: %s\n", datasetName)
//		fmt.Printf("  Size: OurFormat=%dB, Protobuf=%dB, Ratio=%.2fx\n",
//			ourFormat.EncodedSize, protobuf.EncodedSize, sizeRatio)
//		fmt.Printf("  Encode Speed: OurFormat=%.0f/s, Protobuf=%.0f/s, Ratio=%.2fx\n",
//			ourFormat.EncodeSpeed, protobuf.EncodeSpeed, encodeSpeedRatio)
//		fmt.Printf("  Decode Speed: OurFormat=%.0f/s, Protobuf=%.0f/s, Ratio=%.2fx\n",
//			ourFormat.DecodeSpeed, protobuf.DecodeSpeed, decodeSpeedRatio)
//}
