package rider

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

type createRiderRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type updateLocationRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

type riderResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Phone       string   `json:"phone"`
	Status      string   `json:"status"`
	CurrentLat  *float64 `json:"current_lat,omitempty"`
	CurrentLng  *float64 `json:"current_lng,omitempty"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req createRiderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Phone == "" {
		http.Error(w, "name and phone are required", http.StatusBadRequest)
		return
	}

	rider, err := h.service.Register(r.Context(), req.Name, req.Phone)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rider)
}

func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "invalid rider id", http.StatusBadRequest)
		return
	}

	var req updateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateLocation(r.Context(), id, req.Lat, req.Lng); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "location updated"})
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "invalid rider id", http.StatusBadRequest)
		return
	}

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateStatus(r.Context(), id, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "status updated"})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	riders, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(riders)
}

func (h *Handler) ListAvailable(w http.ResponseWriter, r *http.Request) {
	riders, err := h.service.ListAvailable(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(riders)
}