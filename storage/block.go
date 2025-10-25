package storage

import (
	"tsdb/encoding"
	"tsdb/types"
)

type BlockManager struct {
	blockSize int
}

func NewBlockManager(blockSize int) *BlockManager {
	return &BlockManager{
		blockSize: blockSize,
	}
}

func (bm *BlockManager) CreateBlock(points []types.Point) (*types.DataBlock, error) {
	if len(points) == 0 {
		return nil, nil
	}

	minValue := points[0].Value
	maxValue := points[0].Value

	for _, p := range points {
		if p.Value < minValue {
			minValue = p.Value
		}
		if p.Value > maxValue {
			maxValue = p.Value
		}
	}

	compressedTimestamps, compressedValues, err := encoding.CompressPoints(points)
	if err != nil {
		return nil, err
	}

	return &types.DataBlock{
		StartTime:  points[0].Timestamp,
		EndTime:    points[len(points)-1].Timestamp,
		PointCount: int16(len(points)),
		MinValue:   minValue,
		MaxValue:   maxValue,
		Timestamps: compressedTimestamps,
		Values:     compressedValues,
	}, nil
}

func (bm *BlockManager) DecompressBlock(block *types.DataBlock) ([]types.Point, error) {
	return encoding.DecompressPoints(block.Timestamps, block.Values, int(block.PointCount))
}

func (bm *BlockManager) SplitPoints(points []types.Point) [][]types.Point {
	var blocks [][]types.Point

	for i := 0; i < len(points); i += bm.blockSize {
		end := i + bm.blockSize
		if end > len(points) {
			end = len(points)
		}
		blocks = append(blocks, points[i:end])
	}

	return blocks
}
