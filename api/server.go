package api

import (
	"fmt"
	"log"
	"net/http"
	"tsdb/types"
)

type Server struct {
	tsdb   types.TSDB
	server *http.Server
}

func NewServer(tsdb types.TSDB, host string, port int) *Server {
	mux := http.NewServeMux()
	server := &Server{tsdb: tsdb}

	mux.HandleFunc("/write", server.writeHandler)
	mux.HandleFunc("/query", server.queryHandler)
	mux.HandleFunc("/health", server.healthHandler)
	mux.HandleFunc("/series", server.seriesHandler)

	server.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: mux,
	}

	return server
}

func (s *Server) Start() error {
	log.Printf("Starting TSDB server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown() error {
	return s.server.Close()
}
