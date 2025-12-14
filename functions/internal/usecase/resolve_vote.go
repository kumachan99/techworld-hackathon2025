package usecase

import (
	"context"
	"encoding/base64"
	"log/slog"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/domain/repository"
	"github.com/techworld-hackathon/functions/internal/domain/service"
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
	roomRepo       repository.RoomRepository
	playerRepo     repository.PlayerRepository
	policyRepo     repository.PolicyRepository
	imageGenerator service.ImageGenerator
	imageStorage   service.ImageStorage
}

// NewResolveVoteUseCase は ResolveVoteUseCase を作成する
func NewResolveVoteUseCase(
	roomRepo repository.RoomRepository,
	playerRepo repository.PlayerRepository,
	policyRepo repository.PolicyRepository,
	imageGenerator service.ImageGenerator,
	imageStorage service.ImageStorage,
) *ResolveVoteUseCase {
	return &ResolveVoteUseCase{
		roomRepo:       roomRepo,
		playerRepo:     playerRepo,
		policyRepo:     policyRepo,
		imageGenerator: imageGenerator,
		imageStorage:   imageStorage,
	}
}

// Execute は投票を集計し、結果を反映する
// 1. votes を集計して最多得票の政策を決定（同数の場合はランダム）
// 2. master_policies から effects を取得
// 3. cityParams に効果を適用
// 4. isCollapsed をチェック（いずれかのパラメータが 0 以下 or 100 以上）
// 5. lastResult を設定
// 6. status を RESULT に
// 7. ゲーム終了判定: turn >= maxTurns or isCollapsed → FINISHED
// ※ 次のターンの準備（カード引き、投票リセット）は next_turn.go で行う
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
	winningPolicy, err := findPolicy(ctx, room, uc.policyRepo, winningPolicyID)
	if err != nil {
		return nil, err
	}
	if winningPolicy == nil {
		return nil, entity.ErrPolicyNotFound
	}

	// 政策の効果を街に適用
	room.ApplyPolicyEffects(winningPolicy.Effects)

	// 可決された政策を履歴に追加
	room.PassedPolicyIDs = append(room.PassedPolicyIDs, winningPolicy.PolicyID)

	// 投票結果を設定
	room.LastResult = &entity.VoteResult{
		PassedPolicyID:    winningPolicy.PolicyID,
		PassedPolicyTitle: winningPolicy.Title,
		ActualEffects:     winningPolicy.Effects,
		NewsFlash:         winningPolicy.NewsFlash,
		VoteDetails:       room.Votes,
	}

	// 街の画像を生成
	if uc.imageGenerator != nil {
		passedPolicies, err := uc.getPassedPolicies(ctx, room)
		if err != nil {
			slog.Warn("failed to get passed policies for image generation", slog.Any("error", err))
		} else {
			imageResult, err := uc.imageGenerator.GenerateCityImage(ctx, &room.CityParams, passedPolicies)
			if err != nil {
				slog.Warn("failed to generate city image", slog.Any("error", err))
			} else {
				room.LastResult.CityImage = imageResult.Image

				// GCSにアップロードしてsigned URLを取得
				if uc.imageStorage != nil {
					imageData, err := base64.StdEncoding.DecodeString(imageResult.Image)
					if err != nil {
						slog.Warn("failed to decode base64 image", slog.Any("error", err))
					} else {
						signedURL, err := uc.imageStorage.UploadCityImage(ctx, input.RoomID, room.Turn, imageData)
						if err != nil {
							slog.Warn("failed to upload city image to GCS", slog.Any("error", err))
						} else {
							room.LastResult.CityImageURL = signedURL
							slog.Info("city image uploaded to GCS", slog.String("url", signedURL))
						}
					}
				}
			}
		}
	}

	// 結果発表フェーズに移行
	room.Status = entity.RoomStatusResult

	// ゲーム終了判定
	isGameOver := room.IsGameOver()
	if isGameOver {
		room.Finish()
	}

	// 部屋を更新
	if err := uc.roomRepo.Update(ctx, input.RoomID, room); err != nil {
		return nil, err
	}

	return &ResolveVoteOutput{
		Room:       room,
		IsGameOver: isGameOver,
	}, nil
}

// getPassedPolicies は可決された政策のリストを取得する
func (uc *ResolveVoteUseCase) getPassedPolicies(ctx context.Context, room *entity.Room) ([]*entity.MasterPolicy, error) {
	policies := make([]*entity.MasterPolicy, 0, len(room.PassedPolicyIDs))
	for _, policyID := range room.PassedPolicyIDs {
		policy, err := findPolicy(ctx, room, uc.policyRepo, policyID)
		if err != nil {
			return nil, err
		}
		if policy != nil {
			policies = append(policies, policy)
		}
	}
	return policies, nil
}
