package usecase

import (
	"context"
	"math/rand"
	"time"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/domain/repository"
)

// CreateRoomInput は部屋作成の入力
type CreateRoomInput struct {
	UserID      string
	DisplayName string
}

// CreateRoomOutput は部屋作成の出力
type CreateRoomOutput struct {
	RoomID   string
	Status   entity.RoomStatus
	PlayerID string
}

// CreateRoomUseCase は部屋作成のユースケース
// POST /api/rooms
type CreateRoomUseCase struct {
	roomRepo     repository.RoomRepository
	playerRepo   repository.PlayerRepository
	ideologyRepo repository.IdeologyRepository
}

// NewCreateRoomUseCase は CreateRoomUseCase を作成する
func NewCreateRoomUseCase(
	roomRepo repository.RoomRepository,
	playerRepo repository.PlayerRepository,
	ideologyRepo repository.IdeologyRepository,
) *CreateRoomUseCase {
	return &CreateRoomUseCase{
		roomRepo:     roomRepo,
		playerRepo:   playerRepo,
		ideologyRepo: ideologyRepo,
	}
}

// Execute は部屋を作成する
// 1. 新しい部屋を作成
// 2. ホストプレイヤーを追加
// 3. 思想をランダムに割り当て
func (uc *CreateRoomUseCase) Execute(ctx context.Context, input CreateRoomInput) (*CreateRoomOutput, error) {
	// 思想を取得
	ideologies, err := uc.ideologyRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if len(ideologies) == 0 {
		return nil, entity.ErrNoIdeologyAvailable
	}

	// ランダムに思想を選択
	rand.Seed(time.Now().UnixNano())
	selectedIdeology := ideologies[rand.Intn(len(ideologies))]

	// 新しい部屋を作成
	room := entity.NewRoom(input.UserID)

	// 部屋を保存
	roomID, err := uc.roomRepo.Create(ctx, room)
	if err != nil {
		return nil, err
	}

	// ホストプレイヤーを作成
	player := entity.NewPlayer(input.DisplayName, true, &selectedIdeology)

	// プレイヤーを保存
	if err := uc.playerRepo.Create(ctx, roomID, input.UserID, player); err != nil {
		return nil, err
	}

	// votesマップにホストを追加
	room.Votes[input.UserID] = ""
	if err := uc.roomRepo.Update(ctx, roomID, room); err != nil {
		return nil, err
	}

	return &CreateRoomOutput{
		RoomID:   roomID,
		Status:   room.Status,
		PlayerID: input.UserID,
	}, nil
}
