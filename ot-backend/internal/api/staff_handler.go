package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"ot-uat/internal/db"
)

type staffInputRequest struct {
	StaffID     string `json:"staffid"`
	NameEng     string `json:"nameeng"`
	NameChi     string `json:"namechi"`
	DisplayName string `json:"displayname"`
	DomainName  string `json:"domainname"`
	StaffGroup  string `json:"staffgroup"`
}

func (h *OTHandler) Staff(w http.ResponseWriter, r *http.Request) {
	setJSON(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method == http.MethodDelete {
		staffID := strings.TrimSpace(r.URL.Query().Get("staffid"))
		if staffID == "" {
			http.Error(w, "staffid is required", http.StatusBadRequest)
			return
		}
		if err := h.Store.DeleteStaff(r.Context(), staffID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"deleted": staffID})
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	staff, err := h.Store.ListStaff(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"staff": staff})
}

func (h *OTHandler) StaffInput(w http.ResponseWriter, r *http.Request) {
	setJSON(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var req staffInputRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.StaffID = strings.TrimSpace(req.StaffID)
	req.NameEng = strings.TrimSpace(req.NameEng)
	req.NameChi = strings.TrimSpace(req.NameChi)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.DomainName = strings.TrimSpace(req.DomainName)
	req.StaffGroup = strings.TrimSpace(req.StaffGroup)
	if req.StaffID == "" {
		http.Error(w, "staffid is required", http.StatusBadRequest)
		return
	}

	saved, err := h.Store.UpsertStaff(r.Context(), db.Staff{
		StaffID:     req.StaffID,
		NameEng:     req.NameEng,
		NameChi:     req.NameChi,
		DisplayName: req.DisplayName,
		DomainName:  req.DomainName,
		StaffGroup:  req.StaffGroup,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"staff": saved})
}
