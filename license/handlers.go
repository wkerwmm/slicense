package license

import (
	"encoding/json"
	"net/http"
	"time"
)

type VerifyRequest struct {
	Key     string `json:"key"`
	Product string `json:"product"`
}

type VerifyResponse struct {
	Valid       bool       `json:"valid"`
	Reason      string     `json:"reason,omitempty"`
	Key         string     `json:"key,omitempty"`
	Product     string     `json:"product,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	OwnerEmail  string     `json:"owner_email,omitempty"`
	OwnerName   string     `json:"owner_name,omitempty"`
	IsActivated bool       `json:"is_activated,omitempty"`
}

type AuditLogResponse struct {
	Action     string    `json:"action"`
	LicenseKey string    `json:"license_key"`
	Product    string    `json:"product"`
	ChangedAt  time.Time `json:"changed_at"`
	Details    string    `json:"details,omitempty"`
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) VerifyLicense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	license, err := h.service.GetLicense(req.Key, req.Product)
	if err != nil {
		sendResponse(w, VerifyResponse{
			Valid:  false,
			Reason: "License not found",
		})
		return
	}

	if license.ExpiresAt != nil && time.Now().After(*license.ExpiresAt) {
		sendResponse(w, VerifyResponse{
			Valid:  false,
			Reason: "License expired",
		})
		return
	}

	sendResponse(w, VerifyResponse{
		Valid:       true,
		Key:         license.Key,
		Product:     license.Product,
		ExpiresAt:   license.ExpiresAt,
		OwnerEmail:  license.OwnerEmail,
		OwnerName:   license.OwnerName,
		IsActivated: license.IsActivated,
	})
}

func (h *Handler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.service.GetAuditLogs(100)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var response []AuditLogResponse
	for _, log := range logs {
		response = append(response, AuditLogResponse{
			Action:     log.Action,
			LicenseKey: log.LicenseKey,
			Product:    log.Product,
			ChangedAt:  log.ChangedAt,
			Details:    log.Details,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sendResponse(w http.ResponseWriter, resp VerifyResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}