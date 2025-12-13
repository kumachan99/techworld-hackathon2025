package repository

import (
	"context"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
)

// RoomRepository は部屋の永続化を担当するインターフェース
// パス: rooms/{roomId}
type RoomRepository interface {
	// FindByID は指定されたIDの部屋を取得する
	FindByID(ctx context.Context, roomID string) (*entity.Room, error)

	// Update は部屋の情報を更新する
	Update(ctx context.Context, roomID string, room *entity.Room) error
}

// PlayerRepository はプレイヤーの永続化を担当するインターフェース
// パス: rooms/{roomId}/players/{oderId}
type PlayerRepository interface {
	// FindByID は指定されたIDのプレイヤーを取得する
	FindByID(ctx context.Context, roomID, userID string) (*entity.Player, error)

	// FindAllByRoomID は指定された部屋の全プレイヤーを取得する
	FindAllByRoomID(ctx context.Context, roomID string) ([]*entity.Player, error)

	// Update はプレイヤー情報を更新する
	Update(ctx context.Context, roomID, userID string, player *entity.Player) error

	// ClearAllVotes は全プレイヤーの投票状態をリセットする
	ClearAllVotes(ctx context.Context, roomID string) error
}
