package handler

import (
	"encoding/json"
	"log"
	"net/http"

	messaging "github.com/ianwu0915/SettleChat/internal/nats_messaging"
	"github.com/ianwu0915/SettleChat/internal/storage"
)

type RoomHandler struct {
	DB        *storage.PostgresStore
	publisher *messaging.NATSPublisher
	env       string
	EventBus  *messaging.EventBus
}

func NewRoomHandler(store *storage.PostgresStore, publisher *messaging.NATSPublisher, env string, eventBus *messaging.EventBus) *RoomHandler {
	return &RoomHandler{
		DB:        store,
		publisher: publisher,
		env:       env,
		EventBus:  eventBus,
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

// CreateRoom 創建房間:
// 1. 獲取用戶信息
// 2. 創建房間
// 3. 發布用戶加入事件
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

	if h.EventBus != nil {
		if err := h.EventBus.PublishUserJoinedEvent(rid, req.UserID, user.UserName); err != nil {
			log.Printf("Failed to publish New UserJoin event: %v", err)
		} else {
			log.Printf("Published New UserJoin event for %s in room %s", user.UserName, rid)
		}
	}

	// Return {"room_id:" "lkahld "}
	json.NewEncoder(w).Encode(map[string]string{"room_id": rid}) //?
}

// JoinRoom 加入房間:
// Handle HTTP Reqeust for User Join
// 1. 獲取用戶信息
// 2. 發布用戶加入請求事件
func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	var req joinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if h.EventBus != nil {
		if err := h.EventBus.PublishUserJoinedEvent(req.RoomID, req.UserID, req.Username); err != nil {
			log.Printf("Failed to publish New UserJoin event: %v", err)
		} else {
			log.Printf("Published New UserJoin event for %s in room %s", req.Username, req.RoomID)
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

// LeaveRoom 加入房間:
// Handle HTTP Reqeust for User Left
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

	if h.EventBus != nil {
		if err := h.EventBus.PublishUserLeftEvent(req.RoomID, user.ID, user.UserName); err != nil {
			log.Printf("Failed to publish New UserLeft event: %v", err)
		} else {
			log.Printf("Published New UserLeft event for %s in room %s", user.UserName, req.RoomID)
		}
	}


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
