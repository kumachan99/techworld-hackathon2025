package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/newmo-oss/ergo"
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
}

// JoinRoomRequest は部屋参加リクエスト
type JoinRoomRequest struct {
	DisplayName string `json:"displayName"`
}

// LeaveRoomRequest は部屋退出リクエスト
type LeaveRoomRequest struct {
	PlayerID string `json:"playerId"`
}

// ReadyRequest はReady状態トグルリクエスト
type ReadyRequest struct {
	PlayerID string `json:"playerId"`
}

// StartGameRequest はゲーム開始リクエスト
type StartGameRequest struct {
	PlayerID string `json:"playerId"`
}

// VoteRequest は投票リクエスト
type VoteRequest struct {
	PlayerID string `json:"playerId"`
	PolicyID string `json:"policyId"`
}

// PetitionRequest は陳情リクエスト
type PetitionRequest struct {
	PlayerID string `json:"playerId"`
	Text     string `json:"text"`
}

// ============================================================================
// ハンドラー実装
// ============================================================================

// CreateRoom は部屋を作成する
// POST /api/rooms
func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	slog.Info("CreateRoom: リクエスト受信")

	if r.Method != http.MethodPost {
		slog.Warn("CreateRoom: 不正なメソッド", slog.String("method", r.Method))
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// リクエストボディをパース
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("CreateRoom: リクエストボディのパース失敗", slog.Any("error", err))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DisplayName == "" {
		slog.Warn("CreateRoom: displayNameが空")
		respondError(w, http.StatusBadRequest, "displayName is required")
		return
	}

	// プレイヤーIDを生成
	playerID := uuid.New().String()
	slog.Info("CreateRoom: プレイヤーID生成",
		slog.String("playerId", playerID),
		slog.String("displayName", req.DisplayName))

	output, err := h.createRoomUC.Execute(r.Context(), usecase.CreateRoomInput{
		UserID:      playerID,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		slog.Error("CreateRoom: ユースケース実行失敗", slog.Any("error", err))
		handleError(w, err)
		return
	}

	slog.Info("CreateRoom: 部屋作成成功",
		slog.String("roomId", output.RoomID),
		slog.String("playerId", output.PlayerID))
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"roomId":   output.RoomID,
		"status":   output.Status,
		"playerId": output.PlayerID,
	})
}

// JoinRoom は部屋に参加する
// POST /api/rooms/{roomId}/join
func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	slog.Info("JoinRoom: リクエスト受信")

	if r.Method != http.MethodPost {
		slog.Warn("JoinRoom: 不正なメソッド", slog.String("method", r.Method))
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/join")
	if roomID == "" {
		slog.Warn("JoinRoom: roomIdが空")
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	// リクエストボディをパース
	var req JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("JoinRoom: リクエストボディのパース失敗",
			slog.String("roomId", roomID),
			slog.Any("error", err))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DisplayName == "" {
		slog.Warn("JoinRoom: displayNameが空", slog.String("roomId", roomID))
		respondError(w, http.StatusBadRequest, "displayName is required")
		return
	}

	// プレイヤーIDを生成
	playerID := uuid.New().String()
	slog.Info("JoinRoom: 参加処理開始",
		slog.String("roomId", roomID),
		slog.String("playerId", playerID),
		slog.String("displayName", req.DisplayName))

	output, err := h.joinRoomUC.Execute(r.Context(), usecase.JoinRoomInput{
		RoomID:      roomID,
		UserID:      playerID,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		slog.Error("JoinRoom: ユースケース実行失敗",
			slog.String("roomId", roomID),
			slog.Any("error", err))
		handleError(w, err)
		return
	}

	slog.Info("JoinRoom: 参加成功",
		slog.String("roomId", roomID),
		slog.String("playerId", output.PlayerID))
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"playerId": output.PlayerID,
	})
}

// LeaveRoom は部屋から退出する
// POST /api/rooms/{roomId}/leave
func (h *Handler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	slog.Info("LeaveRoom: リクエスト受信")

	if r.Method != http.MethodPost {
		slog.Warn("LeaveRoom: 不正なメソッド", slog.String("method", r.Method))
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/leave")
	if roomID == "" {
		slog.Warn("LeaveRoom: roomIdが空")
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	// リクエストボディをパース
	var req LeaveRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("LeaveRoom: リクエストボディのパース失敗",
			slog.String("roomId", roomID),
			slog.Any("error", err))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PlayerID == "" {
		slog.Warn("LeaveRoom: playerIdが空", slog.String("roomId", roomID))
		respondError(w, http.StatusBadRequest, "playerId is required")
		return
	}

	slog.Info("LeaveRoom: 退出処理開始",
		slog.String("roomId", roomID),
		slog.String("playerId", req.PlayerID))

	output, err := h.leaveRoomUC.Execute(r.Context(), usecase.LeaveRoomInput{
		RoomID: roomID,
		UserID: req.PlayerID,
	})
	if err != nil {
		slog.Error("LeaveRoom: ユースケース実行失敗",
			slog.String("roomId", roomID),
			slog.String("playerId", req.PlayerID),
			slog.Any("error", err))
		handleError(w, err)
		return
	}

	slog.Info("LeaveRoom: 退出成功",
		slog.String("roomId", roomID),
		slog.String("playerId", req.PlayerID))
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": output.Success,
	})
}

// ToggleReady はReady状態をトグルする
// POST /api/rooms/{roomId}/ready
func (h *Handler) ToggleReady(w http.ResponseWriter, r *http.Request) {
	slog.Info("ToggleReady: リクエスト受信")

	if r.Method != http.MethodPost {
		slog.Warn("ToggleReady: 不正なメソッド", slog.String("method", r.Method))
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/ready")
	if roomID == "" {
		slog.Warn("ToggleReady: roomIdが空")
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	// リクエストボディをパース
	var req ReadyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("ToggleReady: リクエストボディのパース失敗",
			slog.String("roomId", roomID),
			slog.Any("error", err))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PlayerID == "" {
		slog.Warn("ToggleReady: playerIdが空", slog.String("roomId", roomID))
		respondError(w, http.StatusBadRequest, "playerId is required")
		return
	}

	slog.Info("ToggleReady: Ready状態トグル開始",
		slog.String("roomId", roomID),
		slog.String("playerId", req.PlayerID))

	output, err := h.toggleReadyUC.Execute(r.Context(), usecase.ToggleReadyInput{
		RoomID: roomID,
		UserID: req.PlayerID,
	})
	if err != nil {
		slog.Error("ToggleReady: ユースケース実行失敗",
			slog.String("roomId", roomID),
			slog.String("playerId", req.PlayerID),
			slog.Any("error", err))
		handleError(w, err)
		return
	}

	slog.Info("ToggleReady: Ready状態変更成功",
		slog.String("roomId", roomID),
		slog.String("playerId", req.PlayerID),
		slog.Bool("isReady", output.IsReady))
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"isReady": output.IsReady,
	})
}

// StartGame はゲーム開始を処理する
// POST /api/rooms/{roomId}/start
func (h *Handler) StartGame(w http.ResponseWriter, r *http.Request) {
	slog.Info("StartGame: リクエスト受信")

	if r.Method != http.MethodPost {
		slog.Warn("StartGame: 不正なメソッド", slog.String("method", r.Method))
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/start")
	if roomID == "" {
		slog.Warn("StartGame: roomIdが空")
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	// リクエストボディをパース
	var req StartGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("StartGame: リクエストボディのパース失敗",
			slog.String("roomId", roomID),
			slog.Any("error", err))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PlayerID == "" {
		slog.Warn("StartGame: playerIdが空", slog.String("roomId", roomID))
		respondError(w, http.StatusBadRequest, "playerId is required")
		return
	}

	slog.Info("StartGame: ゲーム開始処理開始",
		slog.String("roomId", roomID),
		slog.String("playerId", req.PlayerID))

	output, err := h.startGameUC.Execute(r.Context(), usecase.StartGameInput{
		RoomID: roomID,
		UserID: req.PlayerID,
	})
	if err != nil {
		slog.Error("StartGame: ユースケース実行失敗",
			slog.String("roomId", roomID),
			slog.String("playerId", req.PlayerID),
			slog.Any("error", err))
		handleError(w, err)
		return
	}

	slog.Info("StartGame: ゲーム開始成功",
		slog.String("roomId", roomID),
		slog.Int("turn", output.Room.Turn),
		slog.Any("policies", output.Room.CurrentPolicyIDs))
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":           output.Room.Status,
		"turn":             output.Room.Turn,
		"currentPolicyIds": output.Room.CurrentPolicyIDs,
	})
}

// Vote は投票を処理する
// POST /api/rooms/{roomId}/vote
func (h *Handler) Vote(w http.ResponseWriter, r *http.Request) {
	slog.Info("Vote: リクエスト受信")

	if r.Method != http.MethodPost {
		slog.Warn("Vote: 不正なメソッド", slog.String("method", r.Method))
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/vote")
	if roomID == "" {
		slog.Warn("Vote: roomIdが空")
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	// リクエストボディをパース
	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Vote: リクエストボディのパース失敗",
			slog.String("roomId", roomID),
			slog.Any("error", err))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PlayerID == "" {
		slog.Warn("Vote: playerIdが空", slog.String("roomId", roomID))
		respondError(w, http.StatusBadRequest, "playerId is required")
		return
	}
	if req.PolicyID == "" {
		slog.Warn("Vote: policyIdが空",
			slog.String("roomId", roomID),
			slog.String("playerId", req.PlayerID))
		respondError(w, http.StatusBadRequest, "policyId is required")
		return
	}

	slog.Info("Vote: 投票処理開始",
		slog.String("roomId", roomID),
		slog.String("playerId", req.PlayerID),
		slog.String("policyId", req.PolicyID))

	output, err := h.voteUC.Execute(r.Context(), usecase.VoteInput{
		RoomID:   roomID,
		UserID:   req.PlayerID,
		PolicyID: req.PolicyID,
	})
	if err != nil {
		slog.Error("Vote: ユースケース実行失敗",
			slog.String("roomId", roomID),
			slog.String("playerId", req.PlayerID),
			slog.String("policyId", req.PolicyID),
			slog.Any("error", err))
		handleError(w, err)
		return
	}

	slog.Info("Vote: 投票成功",
		slog.String("roomId", roomID),
		slog.String("playerId", req.PlayerID),
		slog.String("policyId", req.PolicyID),
		slog.Bool("allVoted", output.AllVoted),
		slog.Bool("isResolved", output.IsResolved))

	// 自動resolveされた場合はresolve結果も返す
	if output.IsResolved {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success":    output.Success,
			"allVoted":   output.AllVoted,
			"isResolved": output.IsResolved,
			"status":     output.Room.Status,
			"lastResult": output.Room.LastResult,
			"cityParams": output.Room.CityParams,
			"isGameOver": output.IsGameOver,
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":  output.Success,
		"allVoted": output.AllVoted,
	})
}

// ResolveVote は投票集計を処理する
// POST /api/rooms/{roomId}/resolve
// フロントエンドから全員投票完了時に自動でトリガーされる
func (h *Handler) ResolveVote(w http.ResponseWriter, r *http.Request) {
	slog.Info("ResolveVote: リクエスト受信")

	if r.Method != http.MethodPost {
		slog.Warn("ResolveVote: 不正なメソッド", slog.String("method", r.Method))
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/resolve")
	if roomID == "" {
		slog.Warn("ResolveVote: roomIdが空")
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	slog.Info("ResolveVote: 投票集計開始", slog.String("roomId", roomID))

	output, err := h.resolveVoteUC.Execute(r.Context(), usecase.ResolveVoteInput{
		RoomID: roomID,
	})
	if err != nil {
		slog.Error("ResolveVote: ユースケース実行失敗",
			slog.String("roomId", roomID),
			slog.Any("error", err))
		handleError(w, err)
		return
	}

	slog.Info("ResolveVote: 投票集計成功",
		slog.String("roomId", roomID),
		slog.String("passedPolicy", output.Room.LastResult.PassedPolicyTitle),
		slog.Bool("isGameOver", output.IsGameOver))
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":     output.Room.Status,
		"lastResult": output.Room.LastResult,
		"cityParams": output.Room.CityParams,
		"isGameOver": output.IsGameOver,
	})
}

// NextTurn は次ターンへ進む
// POST /api/rooms/{roomId}/next
// フロントエンドから結果確認後に自動でトリガーされる
func (h *Handler) NextTurn(w http.ResponseWriter, r *http.Request) {
	slog.Info("NextTurn: リクエスト受信")

	if r.Method != http.MethodPost {
		slog.Warn("NextTurn: 不正なメソッド", slog.String("method", r.Method))
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/next")
	if roomID == "" {
		slog.Warn("NextTurn: roomIdが空")
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	slog.Info("NextTurn: 次ターン処理開始", slog.String("roomId", roomID))

	output, err := h.nextTurnUC.Execute(r.Context(), usecase.NextTurnInput{
		RoomID: roomID,
	})
	if err != nil {
		slog.Error("NextTurn: ユースケース実行失敗",
			slog.String("roomId", roomID),
			slog.Any("error", err))
		handleError(w, err)
		return
	}

	slog.Info("NextTurn: 次ターン開始成功",
		slog.String("roomId", roomID),
		slog.Int("turn", output.Turn),
		slog.String("status", string(output.Status)))
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": output.Status,
		"turn":   output.Turn,
	})
}

// SubmitPetition はAI陳情を処理する
// POST /api/rooms/{roomId}/petition
func (h *Handler) SubmitPetition(w http.ResponseWriter, r *http.Request) {
	slog.Info("SubmitPetition: リクエスト受信")

	if r.Method != http.MethodPost {
		slog.Warn("SubmitPetition: 不正なメソッド", slog.String("method", r.Method))
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URLからroomIdを取得
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/petition")
	if roomID == "" {
		slog.Warn("SubmitPetition: roomIdが空")
		respondError(w, http.StatusBadRequest, "room ID is required")
		return
	}

	// リクエストボディをパース
	var req PetitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("SubmitPetition: リクエストボディのパース失敗",
			slog.String("roomId", roomID),
			slog.Any("error", err))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PlayerID == "" {
		slog.Warn("SubmitPetition: playerIdが空", slog.String("roomId", roomID))
		respondError(w, http.StatusBadRequest, "playerId is required")
		return
	}
	if req.Text == "" {
		slog.Warn("SubmitPetition: textが空",
			slog.String("roomId", roomID),
			slog.String("playerId", req.PlayerID))
		respondError(w, http.StatusBadRequest, "text is required")
		return
	}

	slog.Info("SubmitPetition: 陳情処理開始",
		slog.String("roomId", roomID),
		slog.String("playerId", req.PlayerID),
		slog.String("text", req.Text))

	output, err := h.submitPetitionUC.Execute(r.Context(), usecase.SubmitPetitionInput{
		RoomID:       roomID,
		PlayerID:     req.PlayerID,
		PetitionText: req.Text,
	})
	if err != nil {
		slog.Error("SubmitPetition: ユースケース実行失敗",
			slog.String("roomId", roomID),
			slog.String("playerId", req.PlayerID),
			slog.Any("error", err))
		handleError(w, err)
		return
	}

	slog.Info("SubmitPetition: 陳情処理完了",
		slog.String("roomId", roomID),
		slog.String("playerId", req.PlayerID),
		slog.Bool("approved", output.Approved),
		slog.String("policyId", output.PolicyID))
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"approved": output.Approved,
		"policyId": output.PolicyID,
		"message":  output.Message,
	})
}

// ============================================================================
// ユーティリティ関数
// ============================================================================

// extractRoomID はURLパスからroomIdを抽出する
// 例: /api/rooms/abc123/start → abc123
func extractRoomID(path, prefix, suffix string) string {
	path = strings.TrimPrefix(path, prefix)
	path = strings.TrimSuffix(path, suffix)
	return path
}

// handleError はエラーに応じた適切なHTTPレスポンスを返す
func handleError(w http.ResponseWriter, err error) {
	// ergo のエラーから属性を取得してログに追加
	var attrs []any
	attrs = append(attrs, slog.Any("error", err))

	// ergo.AttrsAll でエラーに付与された属性を取得
	for attr := range ergo.AttrsAll(err) {
		attrs = append(attrs, attr)
	}

	switch {
	case errors.Is(err, entity.ErrRoomNotFound):
		slog.Warn("handleError: 部屋が見つからない", attrs...)
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, entity.ErrPlayerNotFound):
		slog.Warn("handleError: プレイヤーが見つからない", attrs...)
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, entity.ErrPolicyNotFound):
		slog.Warn("handleError: 政策が見つからない", attrs...)
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, entity.ErrInvalidPhase):
		slog.Warn("handleError: 無効なフェーズ", attrs...)
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrNotEnoughPlayers):
		slog.Warn("handleError: プレイヤー数不足", attrs...)
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrNotAllVoted):
		slog.Warn("handleError: 全員未投票", attrs...)
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrNotAllReady):
		slog.Warn("handleError: 全員未Ready", attrs...)
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrPetitionUsed):
		slog.Warn("handleError: 陳情使用済み", attrs...)
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrGameAlreadyStarted):
		slog.Warn("handleError: ゲーム開始済み", attrs...)
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrPlayerAlreadyInRoom):
		slog.Warn("handleError: 既に部屋に参加済み", attrs...)
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrPlayerNotInRoom):
		slog.Warn("handleError: プレイヤーが部屋にいない", attrs...)
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrRoomFull):
		slog.Warn("handleError: 部屋が満員", attrs...)
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrNoIdeologyAvailable):
		slog.Warn("handleError: 利用可能な思想がない", attrs...)
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, entity.ErrInvalidPolicy):
		slog.Warn("handleError: 無効な政策", attrs...)
		respondError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, entity.ErrNotHost):
		slog.Warn("handleError: ホストではない", attrs...)
		respondError(w, http.StatusForbidden, err.Error())
	default:
		slog.Error("handleError: 内部エラー", attrs...)
		respondError(w, http.StatusInternalServerError, "internal server error")
	}
}

// respondJSON はJSON形式でレスポンスを返す
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
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
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}
