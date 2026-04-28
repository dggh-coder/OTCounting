package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"ot-uat/internal/db"
)

type OTHandler struct {
	Store *db.Store
}

type inputRequest struct {
	OTStaffID string          `json:"otstaffid"`
	Date      string          `json:"date"`
	Period    string          `json:"period"`
	Type      string          `json:"type"`
	StartTime string          `json:"startTime"`
	EndTime   string          `json:"endTime"`
	InputBy   *string         `json:"inputBy"`
	Entries   []db.EntryInput `json:"entries"`
}

type staffInputRequest struct {
	StaffID     string `json:"staffid"`
	NameEng     string `json:"nameeng"`
	NameChi     string `json:"namechi"`
	DisplayName string `json:"displayname"`
	DomainName  string `json:"domainname"`
}

func (h *OTHandler) Input(w http.ResponseWriter, r *http.Request) {
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

	var req inputRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.OTStaffID == "" || req.Date == "" {
		http.Error(w, "otstaffid/date required", http.StatusBadRequest)
		return
	}
	if req.Period != "" && !validPeriod(req.Period) {
		http.Error(w, "period must be 00/01/02", http.StatusBadRequest)
		return
	}

	entries := req.Entries
	if len(entries) == 0 && req.Type != "" {
		entries = []db.EntryInput{{Type: req.Type, StartTime: req.StartTime, EndTime: req.EndTime, InputBy: req.InputBy}}
	}
	if len(entries) == 0 {
		http.Error(w, "at least one entry is required", http.StatusBadRequest)
		return
	}
	for _, e := range entries {
		if !validType(e.Type) || !validTime(e.StartTime) || !validTime(e.EndTime) {
			http.Error(w, "entry requires type(00/01), startTime(HH:MM), endTime(HH:MM)", http.StatusBadRequest)
			return
		}
	}

	saved := make([]db.SavedEntry, 0)
	if req.Period != "" {
		list, err := h.Store.SavePeriodEntries(r.Context(), req.OTStaffID, req.Date, req.Period, entries)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		saved = append(saved, list...)
	} else {
		grouped := map[string][]db.EntryInput{"00": {}, "01": {}, "02": {}}
		for _, e := range entries {
			p, err := periodFromStartTime(e.StartTime)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			grouped[p] = append(grouped[p], e)
		}
		for _, p := range []string{"00", "01", "02"} {
			if len(grouped[p]) == 0 {
				continue
			}
			list, err := h.Store.SavePeriodEntries(r.Context(), req.OTStaffID, req.Date, p, grouped[p])
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			saved = append(saved, list...)
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"saved": saved})
}

func (h *OTHandler) Get(w http.ResponseWriter, r *http.Request) {
	setJSON(w)
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	entries, err := h.Store.GetEntries(r.Context(), q.Get("otstaffid"), q.Get("date"), q.Get("period"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"entries": entries})
}

func (h *OTHandler) Monthly(w http.ResponseWriter, r *http.Request) {
	setJSON(w)
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	year, _ := strconv.Atoi(q.Get("year"))
	month, _ := strconv.Atoi(q.Get("month"))
	rows, err := h.Store.GetMonthlyTotals(r.Context(), year, month)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"totals": rows})
}

func (h *OTHandler) ProcessTexts(w http.ResponseWriter, r *http.Request) {
	setJSON(w)
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	otstaffid := strings.TrimSpace(r.URL.Query().Get("otstaffid"))
	rows, err := h.Store.GetProcessTexts(r.Context(), otstaffid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"rows": rows})
}

func (h *OTHandler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	setJSON(w)
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idRaw := strings.TrimSpace(r.URL.Query().Get("id"))
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "valid id is required", http.StatusBadRequest)
		return
	}
	if err := h.Store.DeleteEntryAndRebuild(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"deleted": id})
}

func (h *OTHandler) Staff(w http.ResponseWriter, r *http.Request) {
	setJSON(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
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
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"staff": saved})
}

func setJSON(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Content-Type", "application/json")
}

func validPeriod(v string) bool { return v == "00" || v == "01" || v == "02" }
func validType(v string) bool   { return v == "00" || v == "01" }

func validTime(v string) bool {
	parts := strings.Split(v, ":")
	if len(parts) != 2 || len(parts[0]) != 2 || len(parts[1]) != 2 {
		return false
	}
	h, errH := strconv.Atoi(parts[0])
	m, errM := strconv.Atoi(parts[1])
	if errH != nil || errM != nil {
		return false
	}
	return h >= 0 && h <= 23 && m >= 0 && m <= 59
}

func periodFromStartTime(hhmm string) (string, error) {
	parts := strings.Split(hhmm, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid time format: %s", hhmm)
	}
	h, errH := strconv.Atoi(parts[0])
	m, errM := strconv.Atoi(parts[1])
	if errH != nil || errM != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return "", fmt.Errorf("invalid time format: %s", hhmm)
	}
	total := h*60 + m
	if total >= 0 && total < 11*60 {
		return "00", nil
	}
	if total >= 11*60 && total < 15*60 {
		return "01", nil
	}
	return "02", nil
}
