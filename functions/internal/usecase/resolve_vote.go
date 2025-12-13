package usecase

import (
	"context"
	"math/rand"
	"time"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/domain/repository"
)

// ResolveVoteInput は投票集計の入力
type ResolveVoteInput struct {
	RoomID string
}

// ResolveVoteOutput は投票集計の出力
type ResolveVoteOutput struct {
	Room       *entity.Room
	IsGameOver bool
}

// ResolveVoteUseCase は投票集計のユースケース
// POST /api/rooms/{roomId}/resolve
type ResolveVoteUseCase struct {
	roomRepo   repository.RoomRepository
	playerRepo repository.PlayerRepository
	policyRepo repository.PolicyRepository
}

// NewResolveVoteUseCase は ResolveVoteUseCase を作成する
func NewResolveVoteUseCase(
	roomRepo repository.RoomRepository,
	playerRepo repository.PlayerRepository,
	policyRepo repository.PolicyRepository,
) *ResolveVoteUseCase {
	return &ResolveVoteUseCase{
		roomRepo:   roomRepo,
		playerRepo: playerRepo,
		policyRepo: policyRepo,
	}
}

// Execute は投票を集計し、結果を反映する
// 1. votes を集計して最多得票の政策を決定（同数の場合はランダム）
// 2. master_policies から effects を取得
// 3. cityParams に効果を適用
// 4. isCollapsed をチェック（いずれかのパラメータが 0 以下 or 100 以上）
// 5. lastResult を設定
// 6. 次のターンの準備:
//   - deckIds から3枚を currentPolicyIds に移動
//   - votes をリセット
//   - 全プレイヤーの hasVoted, currentVote をリセット
//
// 7. status を RESUL
// 7. status を RESULT に
// 8. ゲーム終了判定: turn >= maxTurns or isCollapsed → FINISHED
func (uc *ResolveVoteUseCase) Execute(ctx context.Context, input ResolveVoteInput) (*ResolveVoteOutput, error) {
	// 部屋を取得
	room, err := uc.roomRepo.FindByID(ctx, input.RoomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, entity.ErrRoomNotFound
	}

	// VOTING状態でないと集計できない
	if room.Status != entity.RoomStatusVoting {
		return nil, entity.ErrInvalidPhase
	}

	// プレイヤー数を取得
	players, err := uc.playerRepo.FindAllByRoomID(ctx, input.RoomID)
	if err != nil {
		return nil, err
	}

	// 全員が投票しているか確認
	if !room.AllPlayersVoted(len(players)) {
		return nil, entity.ErrNotAllVoted
	}

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
		PassedPolicyID:    winningPolicy.ID,
		PassedPolicyTitle: winningPolicy.Title,
		ActualEffects:     winningPolicy.Effects,
		NewsFlash:         winningPolicy.NewsFlash,
		VoteDetails:       room.Votes,
	}

	// 結果発表フェーズに移行
	room.Status = entity.RoomStatusResult

	// ゲーム終了判定
	isGameOver := room.IsGameOver()
	if isGameOver {
		room.Finish()
	} else {
		// 次のターンの準備
		uc.prepareNextTurn(room)
	}

	// 部屋を更新
	if err := uc.roomRepo.Update(ctx, input.RoomID, room); err != nil {
		return nil, err
	}

	// 次のターンに進む場合は投票状態をリセット
	if !isGameOver {
		if err := uc.playerRepo.ClearAllVotes(ctx, input.RoomID); err != nil {
			return nil, err
		}
	}

	return &ResolveVoteOutput{
		Room:       room,
		IsGameOver: isGameOver,
	}, nil
}

// prepareNextTurn は次のターンの準備をする
func (uc *ResolveVoteUseCase) prepareNextTurn(room *entity.Room) {
	// デッキから3枚引く
	currentCount := 3
	if len(room.DeckIDs) < currentCount {
		currentCount = len(room.DeckIDs)
	}

	if currentCount > 0 {
		// シャッフル（同じ順番で出ないように）
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(room.DeckIDs), func(i, j int) {
			room.DeckIDs[i], room.DeckIDs[j] = room.DeckIDs[j], room.DeckIDs[i]
		})

		room.CurrentPolicyIDs = room.DeckIDs[:currentCount]
		room.DeckIDs = room.DeckIDs[currentCount:]
	} else {
		room.CurrentPolicyIDs = []string{}
	}

	// 投票リセット
	for k := range room.Votes {
		room.Votes[k] = ""
	}
}
