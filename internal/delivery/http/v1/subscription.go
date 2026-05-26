package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kgdevgo/subscription-api/internal/domain"
	"github.com/kgdevgo/subscription-api/internal/repository/postgres"
	"github.com/kgdevgo/subscription-api/internal/usecase"
)

type SubscriptionHandler struct {
	useCase domain.SubscriptionUseCase
}

func NewSubscriptionHandler(uc domain.SubscriptionUseCase) *SubscriptionHandler {
	return &SubscriptionHandler{useCase: uc}
}

type errorResponse struct {
	Error string `json:"error"`
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var sub domain.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.useCase.Create(r.Context(), &sub); err != nil {
		if errors.Is(err, usecase.ErrInvalidInput) || errors.Is(err, usecase.ErrDateConflict) {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondWithJSON(w, http.StatusCreated, sub)
}

func (h *SubscriptionHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid uuid")
		return
	}

	sub, err := h.useCase.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "subscription not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, sub)
}

func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid uuid")
		return
	}

	var sub domain.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	sub.ID = id

	if err := h.useCase.Update(r.Context(), &sub); err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "subscription not found")
			return
		}
		if errors.Is(err, usecase.ErrInvalidInput) || errors.Is(err, usecase.ErrDateConflict) {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, sub)
}

func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid uuid")
		return
	}

	if err := h.useCase.Delete(r.Context(), id); err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "subscription not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SubscriptionHandler) CalculateTotal(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var filter domain.Filter

	if userIDStr := q.Get("user_id"); userIDStr != "" {
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid user_id uuid")
			return
		}
		filter.UserID = &uid
	}

	if svcName := q.Get("service_name"); svcName != "" {
		filter.ServiceName = &svcName
	}

	if fromStr := q.Get("from"); fromStr != "" {
		t, err := time.Parse(domain.DateFormat, fromStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid 'from' date, use MM-YYYY")
			return
		}
		filter.FromDate = &domain.CustomDate{Time: t}
	}

	if toStr := q.Get("to"); toStr != "" {
		t, err := time.Parse(domain.DateFormat, toStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid 'to' date, use MM-YYYY")
			return
		}
		filter.ToDate = &domain.CustomDate{Time: t}
	}

	total, err := h.useCase.CalculateTotal(r.Context(), filter)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidInput) {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]int64{"total_price": total})
}
