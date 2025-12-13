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

	// Create は新しい部屋を作成する
	Create(ctx context.Context, room *entity.Room) (string, error)

	// Update は部屋の情報を更新する
	Update(ctx context.Context, roomID string, room *entity.Room) error

	// Delete は部屋を削除する
	Delete(ctx context.Context, roomID string) error
}

// PlayerWithID はプレイヤーとそのIDをセットにした構造体
type PlayerWithID struct {
	UserID string
	Player *entity.Player
}

// PlayerRepository はプレイヤーの永続化を担当するインターフェース
// パス: rooms/{roomId}/players/{userId}
type PlayerRepository interface {
	// FindByID は指定されたIDのプレイヤーを取得する
	FindByID(ctx context.Context, roomID, userID string) (*entity.Player, error)

	// FindAllByRoomID は指定された部屋の全プレイヤーを取得する
	FindAllByRoomID(ctx context.Context, roomID string) ([]*entity.Player, error)

	// FindAllWithIDsByRoomID は指定された部屋の全プレイヤーをIDと共に取得する
	FindAllWithIDsByRoomID(ctx context.Context, roomID string) ([]*PlayerWithID, error)

	// Create はプレイヤーを作成する
	Create(ctx context.Context, roomID, userID string, player *entity.Player) error

	// Update はプレイヤー情報を更新する
	Update(ctx context.Context, roomID, userID string, player *entity.Player) error

	// Delete はプレイヤーを削除する
	Delete(ctx context.Context, roomID, userID string) error

	// ClearAllVotes は全プレイヤーの投票状態をリセットする
	ClearAllVotes(ctx context.Context, roomID string) error

	// CountByRoomID は指定された部屋のプレイヤー数を取得する
	CountByRoomID(ctx context.Context, roomID string) (int, error)
}
