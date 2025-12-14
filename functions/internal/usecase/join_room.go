package usecase

import (
	"context"
	"math/rand"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/domain/repository"
)

// JoinRoomInput は部屋参加の入力
type JoinRoomInput struct {
	RoomID      string
	UserID      string
	DisplayName string
}

// JoinRoomOutput は部屋参加の出力
type JoinRoomOutput struct {
	PlayerID string
}

// JoinRoomUseCase は部屋参加のユースケース
// POST /api/rooms/{roomId}/join
type JoinRoomUseCase struct {
	roomRepo     repository.RoomRepository
	playerRepo   repository.PlayerRepository
	ideologyRepo repository.IdeologyRepository
}

// NewJoinRoomUseCase は JoinRoomUseCase を作成する
func NewJoinRoomUseCase(
	roomRepo repository.RoomRepository,
	playerRepo repository.PlayerRepository,
	ideologyRepo repository.IdeologyRepository,
) *JoinRoomUseCase {
	return &JoinRoomUseCase{
		roomRepo:     roomRepo,
		playerRepo:   playerRepo,
		ideologyRepo: ideologyRepo,
	}
}

// Execute は部屋に参加する
// 1. ルームの存在・状態確認（LOBBYのみ参加可）
// 2. 既に参加済みでないか確認
// 3. 未使用の思想からランダムに割り当て
// 4. プレイヤーを追加
// 5. votesに追加
func (uc *JoinRoomUseCase) Execute(ctx context.Context, input JoinRoomInput) (*JoinRoomOutput, error) {
	// 部屋を取得
	room, err := uc.roomRepo.FindByID(ctx, input.RoomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, entity.ErrRoomNotFound
	}

	// LOBBY状態でないと参加できない
	if room.Status != entity.RoomStatusLobby {
		return nil, entity.ErrGameAlreadyStarted
	}

	// 既に参加済みかチェック
	existingPlayer, err := uc.playerRepo.FindByID(ctx, input.RoomID, input.UserID)
	if err != nil {
		return nil, err
	}
	if existingPlayer != nil {
		return nil, entity.ErrPlayerAlreadyInRoom
	}

	// 現在のプレイヤー一覧を取得
	players, err := uc.playerRepo.FindAllByRoomID(ctx, input.RoomID)
	if err != nil {
		return nil, err
	}

	// プレイヤー上限チェック（最大4人）
	const maxPlayers = 4
	if len(players) >= maxPlayers {
		return nil, entity.ErrRoomFull
	}

	// 全思想を取得
	allIdeologies, err := uc.ideologyRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// 使用済み思想IDを収集
	usedIdeologyIDs := make(map[string]bool)
	for _, p := range players {
		if p.Ideology != nil {
			usedIdeologyIDs[p.Ideology.IdeologyID] = true
		}
	}

	// 未使用の思想を収集
	var availableIdeologies []entity.MasterIdeology
	for _, ideology := range allIdeologies {
		if !usedIdeologyIDs[ideology.IdeologyID] {
			availableIdeologies = append(availableIdeologies, ideology)
		}
	}

	// 思想が足りない
	if len(availableIdeologies) == 0 {
		return nil, entity.ErrRoomFull
	}

	// ランダムに思想を選択
	selectedIdeology := availableIdeologies[rand.Intn(len(availableIdeologies))]

	// プレイヤーを作成
	player := entity.NewPlayer(input.DisplayName, false, &selectedIdeology)

	// プレイヤーを保存
	if err := uc.playerRepo.Create(ctx, input.RoomID, input.UserID, player); err != nil {
		return nil, err
	}

	// votesマップにプレイヤーを追加
	room.Votes[input.UserID] = ""
	if err := uc.roomRepo.Update(ctx, input.RoomID, room); err != nil {
		return nil, err
	}

	return &JoinRoomOutput{
		PlayerID: input.UserID,
	}, nil
}
