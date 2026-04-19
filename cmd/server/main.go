package main

import (
	"log"
	"net/http"

	"ot-uat/internal/api"
	"ot-uat/internal/engine"
	"ot-uat/internal/service"
	"ot-uat/internal/web"
)

func main() {
	pageHandler, err := web.NewPageHandler("templates/index.html")
	if err != nil {
		log.Fatalf("failed to load template: %v", err)
	}

	calcService := service.NewCalculateService(engine.NewCalculator())
	calcHandler := &api.CalculateHandler{Service: calcService}

	mux := http.NewServeMux()
	mux.Handle("/", pageHandler)
	mux.Handle("/api/calculate", calcHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	addr := ":8080"
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
