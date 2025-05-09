package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ianwu0915/SettleChat/internal/storage"
)

type RoomHandler struct {
	DB *storage.PostgresStore
}

func NewRoomHandler(store *storage.PostgresStore) *RoomHandler {
	return &RoomHandler{DB: store}
}

type createRoomRequest struct {
	RoomName string `json:"room_name"`
	UserID   string `json:"user_id"`
}

type joinRoomRequest struct {
	RoomID string `json:"room_id"`
	UserID string `json:"user_id"`
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	rid, err := h.DB.CreateRoom(context.Background(), req.RoomName, req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.DB.AddUserToRoom(context.Background(), req.UserID, rid)
	json.NewEncoder(w).Encode(map[string]string{"room_id": rid})
}

func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	var req joinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.DB.AddUserToRoom(context.Background(), req.UserID, req.RoomID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *RoomHandler) GetUserRooms(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "missing user_id", http.StatusBadRequest)
		return
	}
	rooms, err := h.DB.GetUserRooms(context.Background(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(rooms)
}
