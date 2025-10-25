package types

// Point - точка данных (семпл)
type Point struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// SeriesIdentifier - идентификатор временного ряда
type SeriesIdentifier struct {
	Metric string            `json:"metric"`
	Tags   map[string]string `json:"tags"`
}

// TimeRange - временной диапазон
type TimeRange struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

// Query - запрос на чтение
type Query struct {
	Metric    string            `json:"metric"`
	Tags      map[string]string `json:"tags"`
	TimeRange TimeRange         `json:"time_range"`
}

// WriteRequest - запрос на запись
type WriteRequest struct {
	Series []SeriesData `json:"series"`
}

// SeriesData - данные одного ряда
type SeriesData struct {
	SeriesID SeriesIdentifier `json:"series_id"`
	Points   []Point          `json:"points"`
}

// QueryResult - результат запроса
type QueryResult struct {
	Series []SeriesData `json:"series"`
}

// SeriesMetadata - метаданные ряда
type SeriesMetadata struct {
	SeriesID    SeriesIdentifier `json:"series_id"`
	FilePath    string           `json:"file_path"`
	BlockCount  int32            `json:"block_count"`
	TotalPoints int64            `json:"total_points"`
	StartTime   int64            `json:"start_time"`
	EndTime     int64            `json:"end_time"`
	MinValue    float64          `json:"min_value"`
	MaxValue    float64          `json:"max_value"`
	CreatedAt   int64            `json:"created_at"`
}

// DataBlock - блок сжатых данных
type DataBlock struct {
	StartTime  int64   `json:"start_time"`
	EndTime    int64   `json:"end_time"`
	PointCount int16   `json:"point_count"`
	MinValue   float64 `json:"min_value"`
	MaxValue   float64 `json:"max_value"`
	Timestamps []byte  `json:"timestamps"`
	Values     []byte  `json:"values"`
}

// BlockHeader - заголовок блока данных
type BlockHeader struct {
	StartTime  int64
	EndTime    int64
	PointCount int16
	MinValue   float64
	MaxValue   float64
	TsSize     int32
	ValueSize  int32
}

// WriteData данные для записи в WAL
type WriteData struct {
	Series []SeriesData `json:"series"`
}

// DeleteData данные для удаления в WAL
type DeleteData struct {
	SeriesID SeriesIdentifier `json:"series_id"`
	Start    int64            `json:"start"`
	End      int64            `json:"end"`
}
