package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"ot-uat/internal/api"
	"ot-uat/internal/db"
	"ot-uat/internal/engine"
	"ot-uat/internal/service"
)

func main() {
	dsn, err := databaseURLFromEnv()
	if err != nil {
		log.Fatalf("failed to read db config: %v", err)
	}
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
	otHandler := &api.OTHandler{Store: store}

	mux := http.NewServeMux()
	mux.Handle("/api/calculate", calcHandler)
	mux.HandleFunc("/api/ot/input", otHandler.Input)
	mux.HandleFunc("/api/ot/entries", otHandler.Get)
	mux.HandleFunc("/api/ot/monthly", otHandler.Monthly)
	mux.HandleFunc("/api/staff", otHandler.Staff)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	addr := ":8080"
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func databaseURLFromEnv() (string, error) {
	host := envOr("DB_HOST", "opengauss")
	port := envOr("DB_PORT", "5432")
	name := envOr("DB_NAME", "postgres")
	user := envOr("DB_USER", "ot_user")
	password := os.Getenv("DB_PASSWORD")

	if password == "" {
		passFile := os.Getenv("DB_PASSWORD_FILE")
		if passFile == "" {
			return "", fmt.Errorf("set DB_PASSWORD or DB_PASSWORD_FILE")
		}
		b, err := os.ReadFile(passFile)
		if err != nil {
			return "", fmt.Errorf("read DB_PASSWORD_FILE: %w", err)
		}
		password = strings.TrimSpace(string(b))
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, name), nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
