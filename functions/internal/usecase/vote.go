package usecase

import (
	"context"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/domain/repository"
)

// VoteInput は投票の入力
type VoteInput struct {
	RoomID   string
	UserID   string
	PolicyID string
}

// VoteOutput は投票の出力
type VoteOutput struct {
	Success bool
}

// VoteUseCase は投票のユースケース
// POST /api/rooms/{roomId}/vote
type VoteUseCase struct {
	roomRepo   repository.RoomRepository
	playerRepo repository.PlayerRepository
}

// NewVoteUseCase は VoteUseCase を作成する
func NewVoteUseCase(
	roomRepo repository.RoomRepository,
	playerRepo repository.PlayerRepository,
) *VoteUseCase {
	return &VoteUseCase{
		roomRepo:   roomRepo,
		playerRepo: playerRepo,
	}
}

// Execute は投票を行う
// 1. VOTING状態であることを確認
// 2. 有効な政策IDであることを確認（currentPolicyIdsに含まれる）
// 3. プレイヤーのcurrentVoteを更新
// 4. Roomのvotesを更新
func (uc *VoteUseCase) Execute(ctx context.Context, input VoteInput) (*VoteOutput, error) {
	// 部屋を取得
	room, err := uc.roomRepo.FindByID(ctx, input.RoomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, entity.ErrRoomNotFound
	}

	// VOTING状態でないと投票できない
	if room.Status != entity.RoomStatusVoting {
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

	// 有効な政策IDかチェック
	validPolicy := false
	for _, policyID := range room.CurrentPolicyIDs {
		if policyID == input.PolicyID {
			validPolicy = true
			break
		}
	}
	if !validPolicy {
		return nil, entity.ErrInvalidPolicy
	}

	// プレイヤーの投票を更新
	player.CurrentVote = input.PolicyID
	if err := uc.playerRepo.Update(ctx, input.RoomID, input.UserID, player); err != nil {
		return nil, err
	}

	// Roomのvotesを更新
	room.Votes[input.UserID] = input.PolicyID
	if err := uc.roomRepo.Update(ctx, input.RoomID, room); err != nil {
		return nil, err
	}

	return &VoteOutput{
		Success: true,
	}, nil
}
