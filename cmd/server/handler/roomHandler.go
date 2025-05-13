package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ianwu0915/SettleChat/internal/messaging"
	"github.com/ianwu0915/SettleChat/internal/storage"
)

type RoomHandler struct {
	DB        *storage.PostgresStore
	publisher *messaging.NATSPublisher
	env       string
}

func NewRoomHandler(store *storage.PostgresStore, publisher *messaging.NATSPublisher, env string) *RoomHandler {
	return &RoomHandler{
		DB:        store,
		publisher: publisher,
		env:       env,
	}
}

type createRoomRequest struct {
	RoomName string `json:"room_name"`
	UserID   string `json:"user_id"`
}

type joinRoomRequest struct {
	RoomID   string `json:"room_id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

type leaveRoomRequest struct {
	RoomID string `json:"room_id"`
	UserID string `json:"user_id"`
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// 獲取用戶信息
	user, err := h.DB.GetUserByID(r.Context(), req.UserID)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 創建房間
	rid, err := h.DB.CreateRoom(r.Context(), req.RoomName, req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 發布用戶加入事件
	if err := h.publisher.PublishUserJoined(rid, req.UserID, user.UserName); err != nil {
		log.Printf("Failed to publish user joined event: %v", err)
	}

	json.NewEncoder(w).Encode(map[string]string{"room_id": rid})
}

func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	var req joinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// 移除用戶加入事件的發布，因為它已在 Room.AddClient 中處理

	w.WriteHeader(http.StatusOK)
}

func (h *RoomHandler) GetUserRooms(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "missing user_id", http.StatusBadRequest)
		return
	}

	rooms, err := h.DB.GetUserRooms(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(rooms)
}

func (h *RoomHandler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	var req leaveRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// 獲取用戶信息
	user, err := h.DB.GetUserByID(r.Context(), req.UserID)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 發布用戶離開事件
	if err := h.publisher.PublishUserLeft(req.RoomID, req.UserID, user.UserName); err != nil {
		log.Printf("Failed to publish user left event: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
