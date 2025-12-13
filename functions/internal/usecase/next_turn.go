package usecase

import (
	"context"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/domain/repository"
)

// NextTurnInput は次ターンの入力
type NextTurnInput struct {
	RoomID string
}

// NextTurnOutput は次ターンの出力
type NextTurnOutput struct {
	Status entity.RoomStatus
	Turn   int
}

// NextTurnUseCase は次ターンへ進むユースケース
// POST /api/rooms/{roomId}/next
type NextTurnUseCase struct {
	roomRepo   repository.RoomRepository
	playerRepo repository.PlayerRepository
}

// NewNextTurnUseCase は NextTurnUseCase を作成する
func NewNextTurnUseCase(
	roomRepo repository.RoomRepository,
	playerRepo repository.PlayerRepository,
) *NextTurnUseCase {
	return &NextTurnUseCase{
		roomRepo:   roomRepo,
		playerRepo: playerRepo,
	}
}

// Execute は次のターンに進める
// フロントエンドから自動でトリガーされる（ホストチェックなし）
// 1. RESULT状態であることを確認
// 2. turnをインクリメント
// 3. statusをVOTINGに
// 4. 次の3枚の政策をセット
// 5. votesをリセット
func (uc *NextTurnUseCase) Execute(ctx context.Context, input NextTurnInput) (*NextTurnOutput, error) {
	// 部屋を取得
	room, err := uc.roomRepo.FindByID(ctx, input.RoomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, entity.ErrRoomNotFound
	}

	// RESULT状態でないと次ターンに進めない
	if room.Status != entity.RoomStatusResult {
		return nil, entity.ErrInvalidPhase
	}

	// 次の3枚の政策をセット
	currentCount := 3
	if len(room.DeckIDs) < currentCount {
		currentCount = len(room.DeckIDs)
	}
	room.CurrentPolicyIDs = room.DeckIDs[:currentCount]
	room.DeckIDs = room.DeckIDs[currentCount:]

	// turnをインクリメント
	room.Turn++

	// votesをリセット
	for userID := range room.Votes {
		room.Votes[userID] = ""
	}

	// statusをVOTINGに
	room.Status = entity.RoomStatusVoting
	room.LastResult = nil

	// 部屋を更新
	if err := uc.roomRepo.Update(ctx, input.RoomID, room); err != nil {
		return nil, err
	}

	// 全プレイヤーの投票状態をリセット
	if err := uc.playerRepo.ClearAllVotes(ctx, input.RoomID); err != nil {
		return nil, err
	}

	return &NextTurnOutput{
		Status: room.Status,
		Turn:   room.Turn,
	}, nil
}
