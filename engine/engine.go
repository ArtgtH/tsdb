package engine

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"tsdb/index"
	"tsdb/storage"
	"tsdb/types"
	"tsdb/wal"
)

type TSDBEngine struct {
	dataDir       string
	indexManager  *index.IndexManager
	fileManager   *storage.FileManager
	wal           *wal.WAL
	blockSize     int
	activeWriters map[string]*SeriesWriter
	writersMutex  sync.RWMutex
	initialized   bool
}

func NewTSDBEngine(dataDir string, blockSize int) (*TSDBEngine, error) {
	if err := os.MkdirAll(filepath.Join(dataDir, "metrics"), 0755); err != nil {
		return nil, err
	}

	wal, err := wal.NewWAL(filepath.Join(dataDir, "wal"), 64*1024*1024)
	if err != nil {
		return nil, err
	}

	engine := &TSDBEngine{
		dataDir:       dataDir,
		indexManager:  index.NewIndexManager(dataDir),
		fileManager:   storage.NewFileManager(dataDir),
		wal:           wal,
		blockSize:     blockSize,
		activeWriters: make(map[string]*SeriesWriter),
	}

	if err := engine.recoverFromWAL(); err != nil {
		return nil, err
	}

	if err := engine.indexManager.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: could not load index: %v", err)
	}

	if err := engine.restoreWriters(); err != nil {
		return nil, err
	}

	log.Printf("TSDB initialized. Series in index: %d", len(engine.indexManager.GetAllSeries()))
	engine.initialized = true
	return engine, nil
}

func (e *TSDBEngine) Write(request types.WriteRequest) error {
	log.Printf("Writing %d series", len(request.Series))

	walData := types.WriteData{Series: request.Series}
	if err := e.wal.Write("write", walData); err != nil {
		return err
	}

	for _, seriesData := range request.Series {
		seriesHash := e.indexManager.HashSeries(seriesData.SeriesID)

		e.writersMutex.Lock()
		writer, exists := e.activeWriters[seriesHash]
		if !exists {
			metadata, file, err := e.createNewSeries(seriesData.SeriesID)
			if err != nil {
				e.writersMutex.Unlock()
				return err
			}

			writer = NewSeriesWriter(metadata, file, e.blockSize)
			e.activeWriters[seriesHash] = writer
			e.indexManager.AddSeries(metadata)
			log.Printf("Created new series: %s with tags %v", seriesData.SeriesID.Metric, seriesData.SeriesID.Tags)
		}

		if err := writer.WritePoints(seriesData.Points); err != nil {
			e.writersMutex.Unlock()
			return err
		}
		e.writersMutex.Unlock()
	}

	if err := e.indexManager.Save(); err != nil {
		log.Printf("Failed to save index: %v", err)
	}

	return nil
}

func (e *TSDBEngine) Read(query types.Query) (types.QueryResult, error) {
	log.Printf("Query: metric=%s, tags=%v, start=%d, end=%d",
		query.Metric, query.Tags, query.TimeRange.Start, query.TimeRange.End)

	seriesList := e.FindSeries(query.Metric, query.Tags)
	log.Printf("Found %d series matching the query", len(seriesList))

	result := types.QueryResult{
		Series: make([]types.SeriesData, 0, len(seriesList)),
	}

	for i, seriesID := range seriesList {
		log.Printf("Reading series %d: %s %v", i, seriesID.Metric, seriesID.Tags)
		points, err := e.readPointsFromSeries(seriesID, query.TimeRange.Start, query.TimeRange.End)
		if err != nil {
			return result, err
		}

		log.Printf("Series %d has %d points", i, len(points))
		if len(points) > 0 {
			result.Series = append(result.Series, types.SeriesData{
				SeriesID: seriesID,
				Points:   points,
			})
		}
	}

	log.Printf("Query result: %d series with data", len(result.Series))
	return result, nil
}

func (e *TSDBEngine) FindSeries(metric string, tags map[string]string) []types.SeriesIdentifier {
	return e.indexManager.FindSeries(metric, tags)
}

func (e *TSDBEngine) Flush() error {
	e.writersMutex.Lock()
	defer e.writersMutex.Unlock()

	for _, writer := range e.activeWriters {
		if err := writer.Flush(); err != nil {
			return err
		}
	}

	if err := e.indexManager.Save(); err != nil {
		return err
	}

	return nil
}

func (e *TSDBEngine) Close() error {
	if err := e.Flush(); err != nil {
		return err
	}

	e.writersMutex.Lock()
	defer e.writersMutex.Unlock()

	for _, writer := range e.activeWriters {
		if err := writer.Close(); err != nil {
			log.Printf("Error closing writer: %v", err)
		}
	}
	e.activeWriters = make(map[string]*SeriesWriter)

	return e.wal.Close()
}

func (e *TSDBEngine) recoverFromWAL() error {
	return e.wal.Read(func(recordType string, data []byte) error {
		switch recordType {
		case "write":
			var writeData types.WriteData
			if err := json.Unmarshal(data, &writeData); err != nil {
				return err
			}

			for _, seriesData := range writeData.Series {
				seriesHash := e.indexManager.HashSeries(seriesData.SeriesID)

				e.writersMutex.Lock()
				writer, exists := e.activeWriters[seriesHash]
				if !exists {
					metadata, file, err := e.createNewSeries(seriesData.SeriesID)
					if err != nil {
						e.writersMutex.Unlock()
						return err
					}

					writer = NewSeriesWriter(metadata, file, e.blockSize)
					e.activeWriters[seriesHash] = writer
					e.indexManager.AddSeries(metadata)
				}

				if err := writer.WritePoints(seriesData.Points); err != nil {
					e.writersMutex.Unlock()
					return err
				}
				e.writersMutex.Unlock()
			}
		}
		return nil
	})
}

func (e *TSDBEngine) createNewSeries(seriesID types.SeriesIdentifier) (*types.SeriesMetadata, *os.File, error) {
	filePath, file, err := e.fileManager.CreateSeriesFile(seriesID.Metric, seriesID.Tags)
	if err != nil {
		return nil, nil, err
	}

	metadata := &types.SeriesMetadata{
		SeriesID: seriesID,
		FilePath: filePath,
	}

	return metadata, file, nil
}

func (e *TSDBEngine) readPointsFromSeries(seriesID types.SeriesIdentifier, start, end int64) ([]types.Point, error) {
	metadata, exists := e.indexManager.GetSeries(seriesID)
	if !exists {
		log.Printf("Series not found in index: %s %v", seriesID.Metric, seriesID.Tags)
		return nil, nil
	}

	log.Printf("Reading points for series: %s, file: %s", seriesID.Metric, metadata.FilePath)
	return e.fileManager.ReadPointsFromFile(metadata.FilePath, start, end)
}

func (e *TSDBEngine) restoreWriters() error {
	for seriesHash, metadata := range e.indexManager.GetAllSeries() {
		if !e.fileManager.FileExists(metadata.FilePath) {
			filePath, file, err := e.fileManager.GetOrCreateSeriesFile(metadata.SeriesID.Metric, metadata.SeriesID.Tags)
			if err != nil {
				return err
			}

			metadata.FilePath = filePath
			writer := NewSeriesWriter(metadata, file, e.blockSize)
			e.activeWriters[seriesHash] = writer
			continue
		}

		filePath, file, err := e.fileManager.GetOrCreateSeriesFile(metadata.SeriesID.Metric, metadata.SeriesID.Tags)
		if err != nil {
			return err
		}

		metadata.FilePath = filePath

		writer := NewSeriesWriter(metadata, file, e.blockSize)
		e.activeWriters[seriesHash] = writer
	}

	return nil
}

func (e *TSDBEngine) GetAllSeries() map[string]*types.SeriesMetadata {
	return e.indexManager.GetAllSeries()
}

func (e *TSDBEngine) GetSeriesCount() int {
	return len(e.indexManager.GetAllSeries())
}
