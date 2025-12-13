package usecase

import (
	"context"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/domain/repository"
	"github.com/techworld-hackathon/functions/internal/interface/gateway/ai"
)

// SubmitPetitionInput はAI陳情の入力
type SubmitPetitionInput struct {
	RoomID       string
	PlayerID     string
	PetitionText string
}

// SubmitPetitionOutput はAI陳情の出力
type SubmitPetitionOutput struct {
	Approved bool
	PolicyID string
	Message  string
}

// SubmitPetitionUseCase はAI陳情のユースケース
// POST /api/rooms/{roomId}/petitions
type SubmitPetitionUseCase struct {
	roomRepo   repository.RoomRepository
	playerRepo repository.PlayerRepository
	policyRepo repository.PolicyRepository
	aiClient   *ai.OpenAIClient
}

// NewSubmitPetitionUseCase は SubmitPetitionUseCase を作成する
func NewSubmitPetitionUseCase(
	roomRepo repository.RoomRepository,
	playerRepo repository.PlayerRepository,
	policyRepo repository.PolicyRepository,
	aiClient *ai.OpenAIClient,
) *SubmitPetitionUseCase {
	return &SubmitPetitionUseCase{
		roomRepo:   roomRepo,
		playerRepo: playerRepo,
		policyRepo: policyRepo,
		aiClient:   aiClient,
	}
}

// Execute はAI陳情を実行する
// 1. プレイヤーの isPetitionUsed を確認
// 2. OpenAI API で審査
// 3. 承認なら政策カードを生成して deckIds に追加
// 4. プレイヤーの isPetitionUsed を true に
func (uc *SubmitPetitionUseCase) Execute(ctx context.Context, input SubmitPetitionInput) (*SubmitPetitionOutput, error) {
	// 部屋を取得
	room, err := uc.roomRepo.FindByID(ctx, input.RoomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, entity.ErrRoomNotFound
	}

	// VOTING状態でないと陳情できない
	if room.Status != entity.RoomStatusVoting {
		return nil, entity.ErrInvalidPhase
	}

	// プレイヤーを取得
	player, err := uc.playerRepo.FindByID(ctx, input.RoomID, input.PlayerID)
	if err != nil {
		return nil, err
	}
	if player == nil {
		return nil, entity.ErrPlayerNotFound
	}

	// 陳情を使用済みか確認
	if player.IsPetitionUsed {
		return nil, entity.ErrPetitionUsed
	}

	// AI審査
	result, err := uc.aiClient.ReviewPetition(ctx, input.PetitionText)
	if err != nil {
		return nil, err
	}

	if !result.Approved {
		return &SubmitPetitionOutput{
			Approved: false,
			Message:  "提案は審査の結果、却下されました: " + result.Reason,
		}, nil
	}

	// 承認された場合、政策を保存
	policyID, err := uc.policyRepo.Create(ctx, result.Policy)
	if err != nil {
		return nil, err
	}

	// deckIds に追加
	room.DeckIDs = append(room.DeckIDs, policyID)

	// 部屋を更新
	if err := uc.roomRepo.Update(ctx, input.RoomID, room); err != nil {
		return nil, err
	}

	// プレイヤーの陳情フラグを更新
	player.IsPetitionUsed = true
	if err := uc.playerRepo.Update(ctx, input.RoomID, input.PlayerID, player); err != nil {
		return nil, err
	}

	return &SubmitPetitionOutput{
		Approved: true,
		PolicyID: policyID,
		Message:  "提案が承認されました！次ターン以降の選択肢に追加される可能性があります。",
	}, nil
}
