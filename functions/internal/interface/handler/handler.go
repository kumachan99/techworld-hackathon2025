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
	startGameUC      *usecase.StartGameUseCase
	resolveVoteUC    *usecase.ResolveVoteUseCase
	submitPetitionUC *usecase.SubmitPetitionUseCase
}

// NewHandler は Handler を作成する
func NewHandler(
	startGameUC *usecase.StartGameUseCase,
	resolveVoteUC *usecase.ResolveVoteUseCase,
	submitPetitionUC *usecase.SubmitPetitionUseCase,
) *Handler {
	return &Handler{
		startGameUC:      startGameUC,
		resolveVoteUC:    resolveVoteUC,
		submitPetitionUC: submitPetitionUC,
	}
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

	output, err := h.startGameUC.Execute(r.Context(), usecase.StartGameInput{
		RoomID: roomID,
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

	output, err := h.resolveVoteUC.Execute(r.Context(), usecase.ResolveVoteInput{
		RoomID: roomID,
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

// PetitionRequest は陳情リクエスト
type PetitionRequest struct {
	Text string `json:"text"`
}

// SubmitPetition はAI陳情を処理する
// POST /api/rooms/{roomId}/petitions
func (h *Handler) SubmitPetition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/petitions")
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

	// プレイヤーIDはFirebase AuthのUIDから取得（実際はミドルウェアで検証）
	playerID := r.Header.Get("X-Player-ID")
	if playerID == "" {
		respondError(w, http.StatusUnauthorized, "player ID is required")
		return
	}

	output, err := h.submitPetitionUC.Execute(r.Context(), usecase.SubmitPetitionInput{
		RoomID:       roomID,
		PlayerID:     playerID,
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
	case errors.Is(err, entity.ErrPetitionUsed):
		respondError(w, http.StatusConflict, err.Error())
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
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Player-ID")
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
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Player-ID")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}
