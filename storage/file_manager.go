package storage

import (
	"encoding/binary"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"tsdb/encoding"
	"tsdb/types"
)

type FileManager struct {
	dataDir string
}

func NewFileManager(dataDir string) *FileManager {
	return &FileManager{
		dataDir: dataDir,
	}
}

func (fm *FileManager) CreateSeriesFile(metric string, tags map[string]string) (string, *os.File, error) {
	metricDir := filepath.Join(fm.dataDir, "metrics", metric)
	if err := os.MkdirAll(metricDir, 0755); err != nil {
		return "", nil, err
	}

	filename := fm.generateFilename(tags)
	filePath := filepath.Join(metricDir, filename)

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", nil, err
	}

	return filePath, file, nil
}

func (fm *FileManager) GetOrCreateSeriesFile(metric string, tags map[string]string) (string, *os.File, error) {
	metricDir := filepath.Join(fm.dataDir, "metrics", metric)
	if err := os.MkdirAll(metricDir, 0755); err != nil {
		return "", nil, err
	}

	filename := fm.generateFilename(tags)
	filePath := filepath.Join(metricDir, filename)

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", nil, err
	}

	return filePath, file, nil
}

func (fm *FileManager) OpenSeriesFile(filePath string) (*os.File, error) {
	return os.Open(filePath)
}

func (fm *FileManager) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func (fm *FileManager) GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (fm *FileManager) WriteBlock(file *os.File, block *types.DataBlock) error {
	header := types.BlockHeader{
		StartTime:  block.StartTime,
		EndTime:    block.EndTime,
		PointCount: block.PointCount,
		MinValue:   block.MinValue,
		MaxValue:   block.MaxValue,
		TsSize:     int32(len(block.Timestamps)),
		ValueSize:  int32(len(block.Values)),
	}

	if err := binary.Write(file, binary.LittleEndian, header); err != nil {
		return err
	}

	if _, err := file.Write(block.Timestamps); err != nil {
		return err
	}

	if _, err := file.Write(block.Values); err != nil {
		return err
	}

	return file.Sync()
}

func (fm *FileManager) ReadBlock(file *os.File) (*types.DataBlock, error) {
	var header types.BlockHeader

	if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, err
	}

	timestamps := make([]byte, header.TsSize)
	if _, err := file.Read(timestamps); err != nil {
		return nil, err
	}

	values := make([]byte, header.ValueSize)
	if _, err := file.Read(values); err != nil {
		return nil, err
	}

	return &types.DataBlock{
		StartTime:  header.StartTime,
		EndTime:    header.EndTime,
		PointCount: header.PointCount,
		MinValue:   header.MinValue,
		MaxValue:   header.MaxValue,
		Timestamps: timestamps,
		Values:     values,
	}, nil
}

func (fm *FileManager) ReadPointsFromFile(filePath string, startTime, endTime int64) ([]types.Point, error) {
	log.Printf("Reading points from file: %s, time range: [%d, %d]", filePath, startTime, endTime)

	file, err := fm.OpenSeriesFile(filePath)
	if err != nil {
		log.Printf("Error opening file %s: %v", filePath, err)
		return nil, err
	}
	defer file.Close()

	var allPoints []types.Point
	blockCount := 0

	for {
		block, err := fm.ReadBlock(file)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error reading block from file %s: %v", filePath, err)
			return nil, err
		}

		blockCount++
		log.Printf("Read block %d: start=%d, end=%d, points=%d",
			blockCount, block.StartTime, block.EndTime, block.PointCount)

		if block.EndTime < startTime || block.StartTime > endTime {
			log.Printf("Block %d outside time range, skipping", blockCount)
			continue
		}

		points, err := encoding.DecompressPoints(block.Timestamps, block.Values, int(block.PointCount))
		if err != nil {
			log.Printf("Error decompressing points in block %d: %v", blockCount, err)
			return nil, err
		}

		log.Printf("Decompressed %d points from block %d", len(points), blockCount)

		for _, point := range points {
			if point.Timestamp >= startTime && point.Timestamp <= endTime {
				allPoints = append(allPoints, point)
			}
		}
	}

	log.Printf("Total points read from file %s: %d", filePath, len(allPoints))
	return allPoints, nil
}

func (fm *FileManager) ReadAllPointsFromFile(filePath string) ([]types.Point, error) {
	return fm.ReadPointsFromFile(filePath, 0, 1<<62)
}

func (fm *FileManager) generateFilename(tags map[string]string) string {
	var tagPairs []string
	for k, v := range tags {
		tagPairs = append(tagPairs, k+"_"+v)
	}
	sort.Strings(tagPairs)

	filename := "series_" + strings.Join(tagPairs, "_") + ".tsdb"

	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	filename = strings.ReplaceAll(filename, "*", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "\"", "_")
	filename = strings.ReplaceAll(filename, "<", "_")
	filename = strings.ReplaceAll(filename, ">", "_")
	filename = strings.ReplaceAll(filename, "|", "_")

	return filename
}

func (fm *FileManager) ListSeriesFiles(metric string) ([]string, error) {
	metricDir := filepath.Join(fm.dataDir, "metrics", metric)

	if _, err := os.Stat(metricDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	var files []string

	err := filepath.Walk(metricDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".tsdb") {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func (fm *FileManager) DeleteSeriesFile(filePath string) error {
	return os.Remove(filePath)
}
