package api

import (
	"encoding/json"
	"net/http"

	"ot-uat/internal/engine"
	"ot-uat/internal/service"
)

type CalculateHandler struct {
	Service *service.CalculateService
}

func (h *CalculateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	var input engine.CalculateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	output, err := h.Service.Calculate(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(output)
}
