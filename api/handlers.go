package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"tsdb/engine"
	"tsdb/types"
)

type SeriesEntry struct {
	Series []struct {
		Metric string            `json:"metric"`
		Tags   map[string]string `json:"tags"`
		Points []struct {
			Timestamp int64   `json:"timestamp"`
			Value     float64 `json:"value"`
		} `json:"points"`
	} `json:"series"`
}

func (s *Server) writeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SeriesEntry
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	writeReq := types.WriteRequest{
		Series: make([]types.SeriesData, len(req.Series)),
	}

	for i, series := range req.Series {
		points := make([]types.Point, len(series.Points))
		for j, point := range series.Points {
			points[j] = types.Point{
				Timestamp: point.Timestamp,
				Value:     point.Value,
			}
		}

		writeReq.Series[i] = types.SeriesData{
			SeriesID: types.SeriesIdentifier{
				Metric: series.Metric,
				Tags:   series.Tags,
			},
			Points: points,
		}
	}

	if err := s.tsdb.Write(writeReq); err != nil {
		http.Error(w, "Write failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (s *Server) queryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metric := r.URL.Query().Get("metric")

	if metric == "" {
		http.Error(w, "Missing required parameter: metric", http.StatusBadRequest)
		return
	}

	var (
		start, end int64
		err        error
	)

	startStr := r.URL.Query().Get("start")
	if startStr != "" {
		start, err = strconv.ParseInt(startStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid start time", http.StatusBadRequest)
			return
		}
	} else {
		start = 0
	}

	endStr := r.URL.Query().Get("end")
	if endStr != "" {
		end, err = strconv.ParseInt(endStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid end time", http.StatusBadRequest)
			return
		}
	} else {
		end = 1<<63 - 1
	}

	tags := make(map[string]string)
	for key, values := range r.URL.Query() {
		if key != "metric" && key != "start" && key != "end" {
			if len(values) > 0 {
				tags[key] = values[0]
			}
		}
	}

	query := types.Query{
		Metric: metric,
		Tags:   tags,
		TimeRange: types.TimeRange{
			Start: start,
			End:   end,
		},
	}

	result, err := s.tsdb.Read(query)
	if err != nil {
		http.Error(w, "Query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := SeriesEntry{
		Series: make([]struct {
			Metric string            `json:"metric"`
			Tags   map[string]string `json:"tags"`
			Points []struct {
				Timestamp int64   `json:"timestamp"`
				Value     float64 `json:"value"`
			} `json:"points"`
		}, len(result.Series)),
	}

	for i, series := range result.Series {
		response.Series[i].Metric = series.SeriesID.Metric
		response.Series[i].Tags = series.SeriesID.Tags
		response.Series[i].Points = make([]struct {
			Timestamp int64   `json:"timestamp"`
			Value     float64 `json:"value"`
		}, len(series.Points))

		for j, point := range series.Points {
			response.Series[i].Points[j] = struct {
				Timestamp int64   `json:"timestamp"`
				Value     float64 `json:"value"`
			}{
				Timestamp: point.Timestamp,
				Value:     point.Value,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// seriesHandler - не костыль, а оптимизация :)
func (s *Server) seriesHandler(w http.ResponseWriter, r *http.Request) {
	if engine, ok := s.tsdb.(*engine.TSDBEngine); ok {
		metric := r.URL.Query().Get("metric")
		var series []types.SeriesIdentifier

		if metric != "" {
			series = engine.FindSeries(metric, map[string]string{})
		} else {
			allSeries := engine.GetAllSeries()
			series = make([]types.SeriesIdentifier, 0, len(allSeries))
			for _, metadata := range allSeries {
				series = append(series, metadata.SeriesID)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(series)
	} else {
		http.Error(w, "Not available", http.StatusInternalServerError)
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
