package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ianwu0915/SettleChat/internal/storage"
)

type AuthHandler struct {
	DB *storage.PostgresStore
}

func NewAuthHandler(store *storage.PostgresStore) *AuthHandler {
	return &AuthHandler{DB: store}
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	UserID string `json:"user_id"`
	Message string `json:"message,omitempty"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := h.DB.Register(context.Background(), req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 
	}
	json.NewEncoder(w).Encode(authResponse{UserID: userID})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := h.DB.Login(context.Background(), req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(authResponse{UserID: userID})
}

func (h *AuthHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "missing user_id", http.StatusBadRequest)
		return
	}

	user, err := h.DB.GetUserByID(context.Background(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return	
	}

	json.NewEncoder(w).Encode(user)
}

