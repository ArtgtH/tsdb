package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tsdb/api"
	"tsdb/engine"
)

func main() {
	dataDir := flag.String("data-dir", "./tsdb_data", "Data directory")
	host := flag.String("host", "localhost", "Server host")
	port := flag.Int("port", 8080, "Server port")
	blockSize := flag.Int("block-size", 1000, "Points per block")
	flag.Parse()

	log.Println("Initializing TSDB...")
	tsdb, err := engine.NewTSDBEngine(*dataDir, *blockSize)
	if err != nil {
		log.Fatalf("Failed to create TSDB: %v", err)
	}
	defer tsdb.Close()

	server := api.NewServer(tsdb, *host, *port)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	log.Printf("TSDB server running on http://%s:%d", *host, *port)
	log.Println("Endpoints:")
	log.Println("  POST /write - Write data")
	log.Println("  GET  /query - Query data")
	log.Println("  GET  /health - Health check")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down...")
	server.Shutdown()
}
