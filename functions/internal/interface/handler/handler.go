package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/usecase"
)

// Handler は全てのAPIハンドラーをまとめた構造体
type Handler struct {
	createRoomUC     *usecase.CreateRoomUseCase
	joinRoomUC       *usecase.JoinRoomUseCase
	leaveRoomUC      *usecase.LeaveRoomUseCase
	toggleReadyUC    *usecase.ToggleReadyUseCase
	startGameUC      *usecase.StartGameUseCase
	voteUC           *usecase.VoteUseCase
	resolveVoteUC    *usecase.ResolveVoteUseCase
	nextTurnUC       *usecase.NextTurnUseCase
	submitPetitionUC *usecase.SubmitPetitionUseCase
}

// NewHandler は Handler を作成する
func NewHandler(
	createRoomUC *usecase.CreateRoomUseCase,
	joinRoomUC *usecase.JoinRoomUseCase,
	leaveRoomUC *usecase.LeaveRoomUseCase,
	toggleReadyUC *usecase.ToggleReadyUseCase,
	startGameUC *usecase.StartGameUseCase,
	voteUC *usecase.VoteUseCase,
	resolveVoteUC *usecase.ResolveVoteUseCase,
	nextTurnUC *usecase.NextTurnUseCase,
	submitPetitionUC *usecase.SubmitPetitionUseCase,
) *Handler {
	return &Handler{
		createRoomUC:     createRoomUC,
		joinRoomUC:       joinRoomUC,
		leaveRoomUC:      leaveRoomUC,
		toggleReadyUC:    toggleReadyUC,
		startGameUC:      startGameUC,
		voteUC:           voteUC,
		resolveVoteUC:    resolveVoteUC,
		nextTurnUC:       nextTurnUC,
		submitPetitionUC: submitPetitionUC,
	}
}

// ============================================================================
// リクエスト型
// ============================================================================

// CreateRoomRequest は部屋作成リクエスト
type CreateRoomRequest struct {
	DisplayName string `json:"displayName"`
	PhotoURL    string `json:"photoURL"`
}

// JoinRoomRequest は部屋参加リクエスト
type JoinRoomRequest struct {
	DisplayName string `json:"displayName"`
	PhotoURL    string `json:"photoURL"`
}

// VoteRequest は投票リクエスト
type VoteRequest struct {
	PolicyID string `json:"policyId"`
}

// PetitionRequest は陳情リクエスト
type PetitionRequest struct {
	Text string `json:"text"`
}

// ============================================================================
// ハンドラー実装
// ============================================================================

// CreateRoom は部屋を作成する
// POST /api/rooms
func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// ユーザーIDを取得
	userID := getUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "user ID is required")
		return
	}

	// リクエストボディをパース
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DisplayName == "" {
		respondError(w, http.StatusBadRequest, "displayName is required")
		return
	}

	output, err := h.createRoomUC.Execute(r.Context(), usecase.CreateRoomInput{
		UserID:      userID,
		DisplayName: req.DisplayName,
		PhotoURL:    req.PhotoURL,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"roomId": output.RoomID,
		"status": output.Status,
	})
}

// JoinRoom は部屋に参加する
// POST /api/rooms/{roomId}/join
func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/join")
	if roomID == "" {
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	// ユーザーIDを取得
	userID := getUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "user ID is required")
		return
	}

	// リクエストボディをパース
	var req JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DisplayName == "" {
		respondError(w, http.StatusBadRequest, "displayName is required")
		return
	}

	output, err := h.joinRoomUC.Execute(r.Context(), usecase.JoinRoomInput{
		RoomID:      roomID,
		UserID:      userID,
		DisplayName: req.DisplayName,
		PhotoURL:    req.PhotoURL,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": output.Success,
	})
}

// LeaveRoom は部屋から退出する
// POST /api/rooms/{roomId}/leave
func (h *Handler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/leave")
	if roomID == "" {
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	// ユーザーIDを取得
	userID := getUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "user ID is required")
		return
	}

	output, err := h.leaveRoomUC.Execute(r.Context(), usecase.LeaveRoomInput{
		RoomID: roomID,
		UserID: userID,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": output.Success,
	})
}

// ToggleReady はReady状態をトグルする
// POST /api/rooms/{roomId}/ready
func (h *Handler) ToggleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/ready")
	if roomID == "" {
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	// ユーザーIDを取得
	userID := getUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "user ID is required")
		return
	}

	output, err := h.toggleReadyUC.Execute(r.Context(), usecase.ToggleReadyInput{
		RoomID: roomID,
		UserID: userID,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"isReady": output.IsReady,
	})
}

// StartGame はゲーム開始を処理する
// POST /api/rooms/{roomId}/start
func (h *Handler) StartGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/start")
	if roomID == "" {
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	// ユーザーIDを取得
	userID := getUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "user ID is required")
		return
	}

	output, err := h.startGameUC.Execute(r.Context(), usecase.StartGameInput{
		RoomID: roomID,
		UserID: userID,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":           output.Room.Status,
		"turn":             output.Room.Turn,
		"currentPolicyIds": output.Room.CurrentPolicyIDs,
	})
}

// Vote は投票を処理する
// POST /api/rooms/{roomId}/vote
func (h *Handler) Vote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/vote")
	if roomID == "" {
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	// ユーザーIDを取得
	userID := getUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "user ID is required")
		return
	}

	// リクエストボディをパース
	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PolicyID == "" {
		respondError(w, http.StatusBadRequest, "policyId is required")
		return
	}

	output, err := h.voteUC.Execute(r.Context(), usecase.VoteInput{
		RoomID:   roomID,
		UserID:   userID,
		PolicyID: req.PolicyID,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": output.Success,
	})
}

// ResolveVote は投票集計を処理する
// POST /api/rooms/{roomId}/resolve
func (h *Handler) ResolveVote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/resolve")
	if roomID == "" {
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	// ユーザーIDを取得
	userID := getUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "user ID is required")
		return
	}

	output, err := h.resolveVoteUC.Execute(r.Context(), usecase.ResolveVoteInput{
		RoomID: roomID,
		UserID: userID,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":     output.Room.Status,
		"lastResult": output.Room.LastResult,
		"cityParams": output.Room.CityParams,
		"isGameOver": output.IsGameOver,
	})
}

// NextTurn は次ターンへ進む
// POST /api/rooms/{roomId}/next
func (h *Handler) NextTurn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/next")
	if roomID == "" {
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	// ユーザーIDを取得
	userID := getUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "user ID is required")
		return
	}

	output, err := h.nextTurnUC.Execute(r.Context(), usecase.NextTurnInput{
		RoomID: roomID,
		UserID: userID,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": output.Status,
		"turn":   output.Turn,
	})
}

// SubmitPetition はAI陳情を処理する
// POST /api/rooms/{roomId}/petition
func (h *Handler) SubmitPetition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/petition")
	if roomID == "" {
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	// リクエストボディをパース
	var req PetitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Text == "" {
		respondError(w, http.StatusBadRequest, "text is required")
		return
	}

	// ユーザーIDを取得
	userID := getUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "user ID is required")
		return
	}

	output, err := h.submitPetitionUC.Execute(r.Context(), usecase.SubmitPetitionInput{
		RoomID:       roomID,
		PlayerID:     userID,
		PetitionText: req.Text,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"approved": output.Approved,
		"policyId": output.PolicyID,
		"message":  output.Message,
	})
}

// ============================================================================
// ユーティリティ関数
// ============================================================================

// getUserID はリクエストからユーザーIDを取得する
// X-User-IDヘッダーまたはFirebase AuthのUIDを取得
func getUserID(r *http.Request) string {
	// 開発用: X-User-IDヘッダーから取得
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}
	// 従来の X-Player-ID も互換性のためにサポート
	if userID := r.Header.Get("X-Player-ID"); userID != "" {
		return userID
	}
	// 本番環境ではFirebase Auth ミドルウェアで設定されたコンテキストから取得
	// ここでは簡略化のためヘッダーから取得
	return ""
}

// extractRoomID はURLパスからroomIdを抽出する
// 例: /api/rooms/abc123/start → abc123
func extractRoomID(path, prefix, suffix string) string {
	path = strings.TrimPrefix(path, prefix)
	path = strings.TrimSuffix(path, suffix)
	return path
}

// handleError はエラーに応じた適切なHTTPレスポンスを返す
func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, entity.ErrRoomNotFound):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, entity.ErrPlayerNotFound):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, entity.ErrPolicyNotFound):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, entity.ErrInvalidPhase):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrNotEnoughPlayers):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrNotAllVoted):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrNotAllReady):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrPetitionUsed):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrGameAlreadyStarted):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrPlayerAlreadyInRoom):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrPlayerNotInRoom):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrRoomFull):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrNoIdeologyAvailable):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrInvalidPolicy):
		respondError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, entity.ErrNotHost):
		respondError(w, http.StatusForbidden, err.Error())
	default:
		log.Printf("Internal error: %v", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
	}
}

// respondJSON はJSON形式でレスポンスを返す
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-Player-ID")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError はエラーレスポンスを返す
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{
		"error": message,
	})
}

// HandleCORS はCORSプリフライトリクエストを処理する
func HandleCORS(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-Player-ID")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}
