package types

type Writer interface {
	Write(request WriteRequest) error
	Flush() error
}

type Reader interface {
	Read(query Query) (QueryResult, error)
	FindSeries(metric string, tags map[string]string) []SeriesIdentifier
}

type TSDB interface {
	Writer
	Reader
	Close() error
}
