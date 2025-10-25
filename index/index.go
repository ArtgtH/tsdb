package index

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"sort"
	"tsdb/types"
)

type GlobalIndex struct {
	Series         map[string]*types.SeriesMetadata
	MetricToSeries map[string]map[string]bool
	TagIndex       map[string]map[string]map[string]bool
}

func NewGlobalIndex() *GlobalIndex {
	return &GlobalIndex{
		Series:         make(map[string]*types.SeriesMetadata),
		MetricToSeries: make(map[string]map[string]bool),
		TagIndex:       make(map[string]map[string]map[string]bool),
	}
}

type IndexManager struct {
	index     *GlobalIndex
	indexFile string
}

func NewIndexManager(dataDir string) *IndexManager {
	return &IndexManager{
		index:     NewGlobalIndex(),
		indexFile: filepath.Join(dataDir, "global.index"),
	}
}

func (im *IndexManager) AddSeries(metadata *types.SeriesMetadata) {
	seriesHash := im.HashSeries(metadata.SeriesID)

	im.index.Series[seriesHash] = metadata

	if im.index.MetricToSeries[metadata.SeriesID.Metric] == nil {
		im.index.MetricToSeries[metadata.SeriesID.Metric] = make(map[string]bool)
	}
	im.index.MetricToSeries[metadata.SeriesID.Metric][seriesHash] = true

	for tagKey, tagValue := range metadata.SeriesID.Tags {
		if im.index.TagIndex[tagKey] == nil {
			im.index.TagIndex[tagKey] = make(map[string]map[string]bool)
		}
		if im.index.TagIndex[tagKey][tagValue] == nil {
			im.index.TagIndex[tagKey][tagValue] = make(map[string]bool)
		}
		im.index.TagIndex[tagKey][tagValue][seriesHash] = true
	}
}

func (im *IndexManager) GetSeries(seriesID types.SeriesIdentifier) (*types.SeriesMetadata, bool) {
	seriesHash := im.HashSeries(seriesID)
	metadata, exists := im.index.Series[seriesHash]
	return metadata, exists
}

func (im *IndexManager) GetAllSeries() map[string]*types.SeriesMetadata {
	return im.index.Series
}

func (im *IndexManager) HashSeries(seriesID types.SeriesIdentifier) string {
	h := fnv.New64a()
	h.Write([]byte(seriesID.Metric))

	keys := make([]string, 0, len(seriesID.Tags))
	for k := range seriesID.Tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte(seriesID.Tags[k]))
	}

	return fmt.Sprintf("%x", h.Sum64())
}

func (im *IndexManager) Save() error {
	data, err := json.MarshalIndent(im.index, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := im.indexFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, im.indexFile)
}

func (im *IndexManager) Load() error {
	data, err := os.ReadFile(im.indexFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &im.index)
}

func (im *IndexManager) FindSeries(metric string, tagFilters map[string]string) []types.SeriesIdentifier {
	var result []types.SeriesIdentifier

	seriesHashes, exists := im.index.MetricToSeries[metric]
	if !exists {
		return result
	}

	for seriesHash := range seriesHashes {
		metadata := im.index.Series[seriesHash]

		matches := true
		for filterKey, filterValue := range tagFilters {
			actualValue, exists := metadata.SeriesID.Tags[filterKey]
			if !exists || (filterValue != "*" && actualValue != filterValue) {
				matches = false
				break
			}
		}

		if matches {
			result = append(result, metadata.SeriesID)
		}
	}

	return result
}
