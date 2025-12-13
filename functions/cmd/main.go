package main

import (
	"context"
	"log"
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
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	// Firestoreクライアント初期化
	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Firestore: %v", err)
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

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initializeHandler は依存性を注入してハンドラーを初期化する
func initializeHandler(firestoreClient *firestore.Client) *handler.Handler {
	// Repository
	roomRepo := firestoreGateway.NewRoomRepository(firestoreClient)
	playerRepo := firestoreGateway.NewPlayerRepository(firestoreClient)
	policyRepo := firestoreGateway.NewPolicyRepository(firestoreClient)

	// AI Client
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	aiClient := ai.NewOpenAIClient(openaiAPIKey)

	// UseCase
	startGameUC := usecase.NewStartGameUseCase(roomRepo, playerRepo, policyRepo)
	resolveVoteUC := usecase.NewResolveVoteUseCase(roomRepo, playerRepo, policyRepo)
	submitPetitionUC := usecase.NewSubmitPetitionUseCase(roomRepo, playerRepo, policyRepo, aiClient)

	// Handler
	return handler.NewHandler(startGameUC, resolveVoteUC, submitPetitionUC)
}

// setupRoutes はルーティングを設定する
func setupRoutes(mux *http.ServeMux, h *handler.Handler) {
	// API endpoints
	// POST /api/rooms/{roomId}/start - ゲーム開始
	// POST /api/rooms/{roomId}/resolve - 投票集計
	// POST /api/rooms/{roomId}/petitions - AI陳情
	mux.HandleFunc("/api/rooms/", func(w http.ResponseWriter, r *http.Request) {
		if handler.HandleCORS(w, r) {
			return
		}

		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/start"):
			h.StartGame(w, r)
		case strings.HasSuffix(path, "/resolve"):
			h.ResolveVote(w, r)
		case strings.HasSuffix(path, "/petitions"):
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
