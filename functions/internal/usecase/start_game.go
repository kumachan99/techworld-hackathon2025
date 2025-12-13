package usecase

import (
	"context"
	"math/rand"
	"time"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/domain/repository"
)

// StartGameInput はゲーム開始の入力
type StartGameInput struct {
	RoomID string
	UserID string // ホストチェック用
}

// StartGameOutput はゲーム開始の出力
type StartGameOutput struct {
	Room *entity.Room
}

// StartGameUseCase はゲーム開始のユースケース
// POST /api/rooms/{roomId}/start
type StartGameUseCase struct {
	roomRepo   repository.RoomRepository
	playerRepo repository.PlayerRepository
	policyRepo repository.PolicyRepository
}

// NewStartGameUseCase は StartGameUseCase を作成する
func NewStartGameUseCase(
	roomRepo repository.RoomRepository,
	playerRepo repository.PlayerRepository,
	policyRepo repository.PolicyRepository,
) *StartGameUseCase {
	return &StartGameUseCase{
		roomRepo:   roomRepo,
		playerRepo: playerRepo,
		policyRepo: policyRepo,
	}
}

// Execute はゲームを開始する
// 1. ホストであることを確認
// 2. 全員Readyであることを確認
// 3. 全政策IDを取得してシャッフル → deckIds
// 4. 先頭3枚を currentPolicyIds に
// 5. deckIds から3枚を削除
// 6. status を VOTING に、turn を 1 に
// 7. 全プレイヤーの投票状態をリセット
func (uc *StartGameUseCase) Execute(ctx context.Context, input StartGameInput) (*StartGameOutput, error) {
	// 部屋を取得
	room, err := uc.roomRepo.FindByID(ctx, input.RoomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, entity.ErrRoomNotFound
	}

	// ホストチェック
	if room.HostID != input.UserID {
		return nil, entity.ErrNotHost
	}

	// LOBBY状態でないとスタートできない
	if room.Status != entity.RoomStatusLobby {
		return nil, entity.ErrInvalidPhase
	}

	// プレイヤー数を確認
	players, err := uc.playerRepo.FindAllByRoomID(ctx, input.RoomID)
	if err != nil {
		return nil, err
	}
	if !room.CanStart(len(players)) {
		return nil, entity.ErrNotEnoughPlayers
	}

	// 全員Readyかチェック
	for _, p := range players {
		if !p.IsReady && !p.IsHost { // ホストはReady不要
			return nil, entity.ErrNotAllReady
		}
	}

	// 全政策IDを取得
	allPolicyIDs, err := uc.policyRepo.GetAllIDs(ctx)
	if err != nil {
		return nil, err
	}

	// シャッフル
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allPolicyIDs), func(i, j int) {
		allPolicyIDs[i], allPolicyIDs[j] = allPolicyIDs[j], allPolicyIDs[i]
	})

	// 先頭3枚を currentPolicyIds に
	currentCount := 3
	if len(allPolicyIDs) < currentCount {
		currentCount = len(allPolicyIDs)
	}

	room.CurrentPolicyIDs = allPolicyIDs[:currentCount]
	room.DeckIDs = allPolicyIDs[currentCount:]

	// Votes マップを初期化（プレイヤーIDをキーに）
	room.Votes = make(map[string]string)
	for _, p := range players {
		room.Votes[p.DisplayName] = "" // displayName は実際にはプレイヤーIDを使うべきだが、設計に合わせる
	}

	// ゲーム開始
	room.Start()

	// 部屋を更新
	if err := uc.roomRepo.Update(ctx, input.RoomID, room); err != nil {
		return nil, err
	}

	// 全プレイヤーの投票状態をリセット
	if err := uc.playerRepo.ClearAllVotes(ctx, input.RoomID); err != nil {
		return nil, err
	}

	return &StartGameOutput{
		Room: room,
	}, nil
}
