package engine

import (
	"os"
	"tsdb/storage"
	"tsdb/types"
)

type SeriesWriter struct {
	metadata     *types.SeriesMetadata
	file         *os.File
	blockBuffer  []types.Point
	blockSize    int
	blockManager *storage.BlockManager
	fileManager  *storage.FileManager
}

func NewSeriesWriter(metadata *types.SeriesMetadata, file *os.File, blockSize int) *SeriesWriter {
	return &SeriesWriter{
		metadata:     metadata,
		file:         file,
		blockBuffer:  make([]types.Point, 0, blockSize),
		blockSize:    blockSize,
		blockManager: storage.NewBlockManager(blockSize),
		fileManager:  storage.NewFileManager(""),
	}
}

func (sw *SeriesWriter) WritePoints(points []types.Point) error {
	sw.blockBuffer = append(sw.blockBuffer, points...)

	if len(sw.blockBuffer) >= sw.blockSize {
		return sw.Flush()
	}

	return sw.Flush()
}

func (sw *SeriesWriter) Flush() error {
	if len(sw.blockBuffer) == 0 {
		return nil
	}

	block, err := sw.blockManager.CreateBlock(sw.blockBuffer)
	if err != nil {
		return err
	}

	if err := sw.fileManager.WriteBlock(sw.file, block); err != nil {
		return err
	}

	sw.updateMetadata(block)
	sw.blockBuffer = sw.blockBuffer[:0]

	return nil
}

func (sw *SeriesWriter) Close() error {
	return sw.Flush()
}

func (sw *SeriesWriter) updateMetadata(block *types.DataBlock) {
	if sw.metadata.TotalPoints == 0 {
		sw.metadata.StartTime = block.StartTime
		sw.metadata.MinValue = block.MinValue
		sw.metadata.MaxValue = block.MaxValue
	} else {
		if block.StartTime < sw.metadata.StartTime {
			sw.metadata.StartTime = block.StartTime
		}
		if block.MinValue < sw.metadata.MinValue {
			sw.metadata.MinValue = block.MinValue
		}
		if block.MaxValue > sw.metadata.MaxValue {
			sw.metadata.MaxValue = block.MaxValue
		}
	}

	sw.metadata.EndTime = block.EndTime
	sw.metadata.TotalPoints += int64(block.PointCount)
	sw.metadata.BlockCount++
}
