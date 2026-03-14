package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"

	"fakeapi/internal/domain"
	apiHTTP "fakeapi/internal/handler/http"
	"fakeapi/internal/port"
	"fakeapi/internal/repository/sqlite"
	"fakeapi/internal/service"
)

func overrideEndpointsWithFile(ctx context.Context, svc port.EndpointService, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var reqs []apiHTTP.EndpointRequest
	err = json.Unmarshal(b, &reqs)
	if err != nil {
		return fmt.Errorf("unmarshal json: %w", err)
	}

	var endpoints []domain.Endpoint
	for _, er := range reqs {
		endpoints = append(endpoints, apiHTTP.NewDomainEndpointFromRequest(er))
	}

	return svc.OverrideAll(ctx, endpoints)
}

func main() {
	dbPath := os.Getenv("FAKEAPI_DB_PATH")
	if dbPath == "" {
		dbPath = "./file.db"
	}

	err := os.MkdirAll(filepath.Dir(dbPath), 0755)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	err = sqlite.InitDB(db)
	if err != nil {
		log.Fatalf("Failed to apply initial DB configuration: %v", err)
	}

	endRepo := sqlite.NewRepository(db)
	endSvc := service.NewEndpointService(endRepo)

	reqRepo := sqlite.NewRequestRepository(db)
	reqSvc := service.NewRequestService(reqRepo)

	ctx := context.Background()

	args := os.Args
	var endpointsFilePath string
	if len(args) >= 2 {
		endpointsFilePath = args[1]
		err := overrideEndpointsWithFile(ctx, endSvc, endpointsFilePath)
		if err != nil {
			log.Errorf("Failed to override endpoints from file: %v", err)
		}
	}

	metaHandler := apiHTTP.NewMetaHandler(endSvc, reqSvc)
	mainHandler := apiHTTP.NewMainHandler(endSvc, reqSvc, metaHandler)

	port := flag.Int("port", 8080, "Port to run the server on")
	flag.Parse()

	fmt.Printf("Running HTTP server on port %d\n", *port)
	http.HandleFunc("/", mainHandler)

	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
