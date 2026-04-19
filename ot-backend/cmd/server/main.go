package main

import (
	"log"
	"net/http"
	"os"

	"ot-uat/internal/api"
	"ot-uat/internal/db"
	"ot-uat/internal/engine"
	"ot-uat/internal/service"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	store, err := db.NewStore(dsn)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	defer store.Close()
	if err := store.InitSchema(); err != nil {
		log.Fatalf("failed to init schema: %v", err)
	}

	calcService := service.NewCalculateService(engine.NewCalculator(), store)
	calcHandler := &api.CalculateHandler{Service: calcService}

	mux := http.NewServeMux()
	mux.Handle("/api/calculate", calcHandler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	addr := ":8080"
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
