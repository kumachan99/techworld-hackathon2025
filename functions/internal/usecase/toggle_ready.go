package usecase

import (
	"context"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/domain/repository"
)

// ToggleReadyInput はReady状態トグルの入力
type ToggleReadyInput struct {
	RoomID string
	UserID string
}

// ToggleReadyOutput はReady状態トグルの出力
type ToggleReadyOutput struct {
	IsReady bool
}

// ToggleReadyUseCase はReady状態トグルのユースケース
// POST /api/rooms/{roomId}/ready
type ToggleReadyUseCase struct {
	roomRepo   repository.RoomRepository
	playerRepo repository.PlayerRepository
}

// NewToggleReadyUseCase は ToggleReadyUseCase を作成する
func NewToggleReadyUseCase(
	roomRepo repository.RoomRepository,
	playerRepo repository.PlayerRepository,
) *ToggleReadyUseCase {
	return &ToggleReadyUseCase{
		roomRepo:   roomRepo,
		playerRepo: playerRepo,
	}
}

// Execute はReady状態をトグルする
// 1. LOBBY状態であることを確認
// 2. isReadyをトグル
func (uc *ToggleReadyUseCase) Execute(ctx context.Context, input ToggleReadyInput) (*ToggleReadyOutput, error) {
	// 部屋を取得
	room, err := uc.roomRepo.FindByID(ctx, input.RoomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, entity.ErrRoomNotFound
	}

	// LOBBY状態でないとReadyできない
	if room.Status != entity.RoomStatusLobby {
		return nil, entity.ErrInvalidPhase
	}

	// プレイヤーを取得
	player, err := uc.playerRepo.FindByID(ctx, input.RoomID, input.UserID)
	if err != nil {
		return nil, err
	}
	if player == nil {
		return nil, entity.ErrPlayerNotInRoom
	}

	// isReadyをトグル
	player.IsReady = !player.IsReady

	// プレイヤー情報を更新
	if err := uc.playerRepo.Update(ctx, input.RoomID, input.UserID, player); err != nil {
		return nil, err
	}

	return &ToggleReadyOutput{
		IsReady: player.IsReady,
	}, nil
}
