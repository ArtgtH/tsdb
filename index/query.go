package index

import "tsdb/types"

type Query struct {
	Metric    string
	Tags      map[string]string
	StartTime int64
	EndTime   int64
}

type QueryResult struct {
	Series types.SeriesIdentifier
	Points []types.Point
}

type QueryEngine struct {
	indexManager *IndexManager
}

func NewQueryEngine(indexManager *IndexManager) *QueryEngine {
	return &QueryEngine{
		indexManager: indexManager,
	}
}

func (qe *QueryEngine) FindSeries(metric string, tagFilters map[string]string) []types.SeriesIdentifier {
	var result []types.SeriesIdentifier

	seriesHashes, hasMetric := qe.indexManager.index.MetricToSeries[metric]
	if !hasMetric {
		return result
	}

	for seriesHash := range seriesHashes {
		metadata := qe.indexManager.index.Series[seriesHash]

		if qe.matchesFilters(metadata.SeriesID, tagFilters) {
			result = append(result, metadata.SeriesID)
		}
	}

	return result
}

func (qe *QueryEngine) matchesFilters(seriesID types.SeriesIdentifier, filters map[string]string) bool {
	for filterKey, filterValue := range filters {
		actualValue, exists := seriesID.Tags[filterKey]

		if !exists {
			return false
		}

		if filterValue != "*" && actualValue != filterValue {
			return false
		}
	}

	return true
}
