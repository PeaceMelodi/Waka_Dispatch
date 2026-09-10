package delivery

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type updateDeliveryStatusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "invalid delivery id", http.StatusBadRequest)
		return
	}

	var req updateDeliveryStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Status != "picked_up" && req.Status != "in_transit" && req.Status != "delivered" && req.Status != "cancelled" {
		http.Error(w, "status must be picked_up, in_transit, delivered, or cancelled", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateStatus(r.Context(), id, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "delivery status updated"})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "invalid delivery id", http.StatusBadRequest)
		return
	}

	delivery, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(delivery)
}

func (h *Handler) ListActive(w http.ResponseWriter, r *http.Request) {
	deliveries, err := h.service.ListActive(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deliveries)
}