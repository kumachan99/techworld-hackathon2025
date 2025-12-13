package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"

	"github.com/techworld-hackathon/functions/internal/interface/gateway/ai"
	firestoreGateway "github.com/techworld-hackathon/functions/internal/interface/gateway/firestore"
	"github.com/techworld-hackathon/functions/internal/interface/handler"
	"github.com/techworld-hackathon/functions/internal/usecase"
)

func main() {
	ctx := context.Background()

	// slogロガーの初期化
	logLevel := slog.LevelInfo
	if os.Getenv("DEBUG") == "true" {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	// Firebase初期化
	// ローカル開発時は FIRESTORE_EMULATOR_HOST が設定されていると自動でエミュレータに接続
	projectID := os.Getenv("GCP_PROJECT")
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	if projectID == "" {
		projectID = "demo-project" // ローカル開発用デフォルト
	}

	conf := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		slog.Error("Failed to initialize Firebase", slog.Any("error", err))
		os.Exit(1)
	}

	// Firestoreクライアント初期化
	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		slog.Error("Failed to initialize Firestore", slog.Any("error", err))
		os.Exit(1)
	}
	defer firestoreClient.Close()

	// 依存性の注入
	h := initializeHandler(firestoreClient)

	// ルーティング設定
	mux := http.NewServeMux()
	setupRoutes(mux, h)

	// サーバー起動
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // ローカルではFirestoreエミュレータが8080を使用
	}

	slog.Info("Server starting", slog.String("port", port))
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("Failed to start server", slog.Any("error", err))
		os.Exit(1)
	}
}

// initializeHandler は依存性を注入してハンドラーを初期化する
func initializeHandler(firestoreClient *firestore.Client) *handler.Handler {
	// Repository
	roomRepo := firestoreGateway.NewRoomRepository(firestoreClient)
	playerRepo := firestoreGateway.NewPlayerRepository(firestoreClient)
	policyRepo := firestoreGateway.NewPolicyRepository(firestoreClient)
	ideologyRepo := firestoreGateway.NewIdeologyRepository(firestoreClient)

	// AI Client
	aiClient := ai.NewSakuraAIClient()

	// UseCase
	createRoomUC := usecase.NewCreateRoomUseCase(roomRepo, playerRepo, ideologyRepo)
	joinRoomUC := usecase.NewJoinRoomUseCase(roomRepo, playerRepo, ideologyRepo)
	leaveRoomUC := usecase.NewLeaveRoomUseCase(roomRepo, playerRepo)
	toggleReadyUC := usecase.NewToggleReadyUseCase(roomRepo, playerRepo)
	startGameUC := usecase.NewStartGameUseCase(roomRepo, playerRepo, policyRepo)
	voteUC := usecase.NewVoteUseCase(roomRepo, playerRepo, policyRepo)
	resolveVoteUC := usecase.NewResolveVoteUseCase(roomRepo, playerRepo, policyRepo)
	nextTurnUC := usecase.NewNextTurnUseCase(roomRepo, playerRepo)
	submitPetitionUC := usecase.NewSubmitPetitionUseCase(roomRepo, playerRepo, policyRepo, aiClient)

	// Handler
	return handler.NewHandler(
		createRoomUC,
		joinRoomUC,
		leaveRoomUC,
		toggleReadyUC,
		startGameUC,
		voteUC,
		resolveVoteUC,
		nextTurnUC,
		submitPetitionUC,
	)
}

// setupRoutes はルーティングを設定する
func setupRoutes(mux *http.ServeMux, h *handler.Handler) {
	// API endpoints
	// POST /api/rooms              - 部屋作成
	// POST /api/rooms/{roomId}/join     - 部屋参加
	// POST /api/rooms/{roomId}/leave    - 部屋退出
	// POST /api/rooms/{roomId}/ready    - Ready状態トグル
	// POST /api/rooms/{roomId}/start    - ゲーム開始
	// POST /api/rooms/{roomId}/vote     - 投票
	// POST /api/rooms/{roomId}/resolve  - 投票集計
	// POST /api/rooms/{roomId}/next     - 次ターンへ
	// POST /api/rooms/{roomId}/petition - AI陳情

	mux.HandleFunc("/api/rooms", func(w http.ResponseWriter, r *http.Request) {
		if handler.HandleCORS(w, r) {
			return
		}
		// /api/rooms のみ（サブパスなし）
		if r.URL.Path == "/api/rooms" {
			h.CreateRoom(w, r)
			return
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("/api/rooms/", func(w http.ResponseWriter, r *http.Request) {
		if handler.HandleCORS(w, r) {
			return
		}

		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/join"):
			h.JoinRoom(w, r)
		case strings.HasSuffix(path, "/leave"):
			h.LeaveRoom(w, r)
		case strings.HasSuffix(path, "/ready"):
			h.ToggleReady(w, r)
		case strings.HasSuffix(path, "/start"):
			h.StartGame(w, r)
		case strings.HasSuffix(path, "/vote"):
			h.Vote(w, r)
		case strings.HasSuffix(path, "/resolve"):
			h.ResolveVote(w, r)
		case strings.HasSuffix(path, "/next"):
			h.NextTurn(w, r)
		case strings.HasSuffix(path, "/petition"):
			h.SubmitPetition(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}
