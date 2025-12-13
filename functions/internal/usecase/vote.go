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
	Success    bool
	AllVoted   bool         // 全員投票済みか
	IsResolved bool         // 自動でresolveされたか
	Room       *entity.Room // resolve後の部屋情報（resolveされた場合のみ）
	IsGameOver bool         // ゲーム終了か
}

// VoteUseCase は投票のユースケース
// POST /api/rooms/{roomId}/vote
type VoteUseCase struct {
	roomRepo   repository.RoomRepository
	playerRepo repository.PlayerRepository
	policyRepo repository.PolicyRepository
}

// NewVoteUseCase は VoteUseCase を作成する
func NewVoteUseCase(
	roomRepo repository.RoomRepository,
	playerRepo repository.PlayerRepository,
	policyRepo repository.PolicyRepository,
) *VoteUseCase {
	return &VoteUseCase{
		roomRepo:   roomRepo,
		playerRepo: playerRepo,
		policyRepo: policyRepo,
	}
}

// Execute は投票を行う
// 1. VOTING状態であることを確認
// 2. 有効な政策IDであることを確認（currentPolicyIdsに含まれる）
// 3. プレイヤーのcurrentVoteを更新
// 4. Roomのvotesを更新
// 5. 全員投票済みなら自動でresolveを実行
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

	// 全プレイヤー数を取得
	players, err := uc.playerRepo.FindAllByRoomID(ctx, input.RoomID)
	if err != nil {
		return nil, err
	}

	// 全員投票済みかチェック
	allVoted := room.AllPlayersVoted(len(players))
	if !allVoted {
		return &VoteOutput{
			Success:  true,
			AllVoted: false,
		}, nil
	}

	// 全員投票済みなら自動でresolveを実行
	return uc.resolveVote(ctx, input.RoomID, room, players)
}

// resolveVote は投票を集計し、結果を反映する（内部メソッド）
func (uc *VoteUseCase) resolveVote(ctx context.Context, roomID string, room *entity.Room, players []*entity.Player) (*VoteOutput, error) {
	// 投票集計
	winningPolicyID := room.CountVotes()

	// 可決された政策を取得
	winningPolicy, err := uc.policyRepo.FindByID(ctx, winningPolicyID)
	if err != nil {
		return nil, err
	}
	if winningPolicy == nil {
		return nil, entity.ErrPolicyNotFound
	}

	// 政策の効果を街に適用
	room.ApplyPolicyEffects(winningPolicy.Effects)

	// 投票結果を設定
	room.LastResult = &entity.VoteResult{
		PassedPolicyID:    winningPolicy.PolicyID,
		PassedPolicyTitle: winningPolicy.Title,
		ActualEffects:     winningPolicy.Effects,
		NewsFlash:         winningPolicy.NewsFlash,
		VoteDetails:       room.Votes,
	}

	// 可決された政策を履歴に追加
	room.PassedPolicyIDs = append(room.PassedPolicyIDs, winningPolicy.PolicyID)

	// 結果発表フェーズに移行
	room.Status = entity.RoomStatusResult

	// ゲーム終了判定
	isGameOver := room.IsGameOver()
	if isGameOver {
		room.Finish()
	}

	// 部屋を更新
	if err := uc.roomRepo.Update(ctx, roomID, room); err != nil {
		return nil, err
	}

	return &VoteOutput{
		Success:    true,
		AllVoted:   true,
		IsResolved: true,
		Room:       room,
		IsGameOver: isGameOver,
	}, nil
}
