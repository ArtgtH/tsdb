package wal

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type WALRecord struct {
	Type      string    `json:"type"` // "write", "delete"
	Timestamp time.Time `json:"timestamp"`
	Data      []byte    `json:"data"`
}

type WAL struct {
	dataDir      string
	currentFile  *os.File
	currentSize  int64
	maxFileSize  int64
	segmentIndex int
	mutex        sync.RWMutex
}

func NewWAL(dataDir string, maxFileSize int64) (*WAL, error) {
	wal := &WAL{
		dataDir:     dataDir,
		maxFileSize: maxFileSize,
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	if err := wal.openOrCreateSegment(); err != nil {
		return nil, err
	}

	return wal, nil
}

func (w *WAL) Write(recordType string, data interface{}) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	record := WALRecord{
		Type:      recordType,
		Timestamp: time.Now(),
		Data:      jsonData,
	}

	recordData, err := json.Marshal(record)
	if err != nil {
		return err
	}

	lengthBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthBuf, uint32(len(recordData)))

	totalSize := int64(len(lengthBuf) + len(recordData))

	if w.currentSize+totalSize >= w.maxFileSize {
		if err := w.rotate(); err != nil {
			return err
		}
	}

	if _, err := w.currentFile.Write(lengthBuf); err != nil {
		return err
	}

	if _, err := w.currentFile.Write(recordData); err != nil {
		return err
	}

	if err := w.currentFile.Sync(); err != nil {
		return err
	}

	w.currentSize += totalSize

	return nil
}

func (w *WAL) Read(handler func(recordType string, data []byte) error) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	currentFile := w.currentFile
	currentSize := w.currentSize
	segmentIndex := w.segmentIndex

	if w.currentFile != nil {
		w.currentFile.Close()
		w.currentFile = nil
	}

	segments, err := w.getSegments()
	if err != nil {
		return err
	}

	for _, segment := range segments {
		file, err := os.Open(segment)
		if err != nil {
			return err
		}

		for {
			lengthBuf := make([]byte, 4)
			_, err := file.Read(lengthBuf)
			if err != nil {
				if err == io.EOF {
					break
				}
				file.Close()
				return err
			}

			length := binary.LittleEndian.Uint32(lengthBuf)
			recordBuf := make([]byte, length)

			if _, err := file.Read(recordBuf); err != nil {
				if err == io.EOF {
					break
				}
				file.Close()
				return err
			}

			var record WALRecord
			if err := json.Unmarshal(recordBuf, &record); err != nil {
				continue
			}

			if err := handler(record.Type, record.Data); err != nil {
				file.Close()
				return err
			}
		}

		file.Close()
	}

	if currentFile != nil {
		filename := filepath.Join(w.dataDir, fmt.Sprintf("segment_%04d.wal", segmentIndex))
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		w.currentFile = file
		w.currentSize = currentSize
		w.segmentIndex = segmentIndex
	} else {
		if err := w.openOrCreateSegment(); err != nil {
			return err
		}
	}

	return nil
}

func (w *WAL) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.currentFile != nil {
		return w.currentFile.Close()
	}
	return nil
}

func (w *WAL) Rotate() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.rotate()
}

func (w *WAL) openOrCreateSegment() error {
	segments, err := w.getSegments()
	if err != nil {
		return err
	}

	if len(segments) > 0 {
		var maxIndex int
		for _, segment := range segments {
			var index int
			_, err := fmt.Sscanf(filepath.Base(segment), "segment_%d.wal", &index)
			if err == nil && index > maxIndex {
				maxIndex = index
			}
		}
		w.segmentIndex = maxIndex + 1
	} else {
		w.segmentIndex = 1
	}

	filename := filepath.Join(w.dataDir, fmt.Sprintf("segment_%04d.wal", w.segmentIndex))
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	w.currentFile = file

	info, err := file.Stat()
	if err != nil {
		return err
	}
	w.currentSize = info.Size()

	return nil
}

func (w *WAL) rotate() error {
	if w.currentFile != nil {
		w.currentFile.Close()
		w.currentFile = nil
	}
	return w.openOrCreateSegment()
}

func (w *WAL) getSegments() ([]string, error) {
	files, err := os.ReadDir(w.dataDir)
	if err != nil {
		return nil, err
	}

	var segments []string
	for _, file := range files {
		if !file.IsDir() {
			segments = append(segments, filepath.Join(w.dataDir, file.Name()))
		}
	}

	sort.Slice(segments, func(i, j int) bool {
		return segments[i] < segments[j]
	})

	return segments, nil
}
