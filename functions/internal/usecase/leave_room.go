package usecase

import (
	"context"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/domain/repository"
)

// LeaveRoomInput は部屋退出の入力
type LeaveRoomInput struct {
	RoomID string
	UserID string
}

// LeaveRoomOutput は部屋退出の出力
type LeaveRoomOutput struct {
	Success bool
}

// LeaveRoomUseCase は部屋退出のユースケース
// POST /api/rooms/{roomId}/leave
type LeaveRoomUseCase struct {
	roomRepo   repository.RoomRepository
	playerRepo repository.PlayerRepository
}

// NewLeaveRoomUseCase は LeaveRoomUseCase を作成する
func NewLeaveRoomUseCase(
	roomRepo repository.RoomRepository,
	playerRepo repository.PlayerRepository,
) *LeaveRoomUseCase {
	return &LeaveRoomUseCase{
		roomRepo:   roomRepo,
		playerRepo: playerRepo,
	}
}

// Execute は部屋から退出する
// 1. プレイヤーを削除
// 2. votesから削除
// 3. ホストが退出した場合、別のプレイヤーをホストに昇格（または部屋を削除）
func (uc *LeaveRoomUseCase) Execute(ctx context.Context, input LeaveRoomInput) (*LeaveRoomOutput, error) {
	// 部屋を取得
	room, err := uc.roomRepo.FindByID(ctx, input.RoomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, entity.ErrRoomNotFound
	}

	// プレイヤーを取得
	player, err := uc.playerRepo.FindByID(ctx, input.RoomID, input.UserID)
	if err != nil {
		return nil, err
	}
	if player == nil {
		return nil, entity.ErrPlayerNotInRoom
	}

	// プレイヤーを削除
	if err := uc.playerRepo.Delete(ctx, input.RoomID, input.UserID); err != nil {
		return nil, err
	}

	// votesから削除
	delete(room.Votes, input.UserID)

	// 残りのプレイヤーを取得（IDと共に）
	remainingPlayers, err := uc.playerRepo.FindAllWithIDsByRoomID(ctx, input.RoomID)
	if err != nil {
		return nil, err
	}

	// プレイヤーがいなくなったら部屋を削除
	if len(remainingPlayers) == 0 {
		if err := uc.roomRepo.Delete(ctx, input.RoomID); err != nil {
			return nil, err
		}
		return &LeaveRoomOutput{Success: true}, nil
	}

	// ホストが退出した場合、別のプレイヤーをホストに昇格
	if player.IsHost {
		// 最初のプレイヤーを新ホストに
		newHostData := remainingPlayers[0]
		newHostData.Player.IsHost = true

		// 新ホストのプレイヤー情報を更新
		if err := uc.playerRepo.Update(ctx, input.RoomID, newHostData.UserID, newHostData.Player); err != nil {
			return nil, err
		}
		room.HostID = newHostData.UserID
	}

	// 部屋を更新
	if err := uc.roomRepo.Update(ctx, input.RoomID, room); err != nil {
		return nil, err
	}

	return &LeaveRoomOutput{
		Success: true,
	}, nil
}
