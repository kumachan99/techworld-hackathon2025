package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/domain/repository"
)

const roomCollection = "rooms"
const playerSubCollection = "players"

// RoomRepository は Firestore を使った RoomRepository の実装
type RoomRepository struct {
	client *firestore.Client
}

// NewRoomRepository は RoomRepository を作成する
func NewRoomRepository(client *firestore.Client) repository.RoomRepository {
	return &RoomRepository{
		client: client,
	}
}

// FindByID は指定されたIDの部屋を取得する
func (r *RoomRepository) FindByID(ctx context.Context, roomID string) (*entity.Room, error) {
	doc, err := r.client.Collection(roomCollection).Doc(roomID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var room entity.Room
	if err := doc.DataTo(&room); err != nil {
		return nil, err
	}
	return &room, nil
}

// Create は新しい部屋を作成する
func (r *RoomRepository) Create(ctx context.Context, room *entity.Room) (string, error) {
	docRef, _, err := r.client.Collection(roomCollection).Add(ctx, room)
	if err != nil {
		return "", err
	}
	return docRef.ID, nil
}

// Update は部屋の情報を更新する
func (r *RoomRepository) Update(ctx context.Context, roomID string, room *entity.Room) error {
	_, err := r.client.Collection(roomCollection).Doc(roomID).Set(ctx, room)
	return err
}

// Delete は部屋を削除する
func (r *RoomRepository) Delete(ctx context.Context, roomID string) error {
	// サブコレクションのプレイヤーも削除
	docs, err := r.client.Collection(roomCollection).Doc(roomID).
		Collection(playerSubCollection).Documents(ctx).GetAll()
	if err != nil {
		return err
	}

	batch := r.client.Batch()
	for _, doc := range docs {
		batch.Delete(doc.Ref)
	}
	// ルームドキュメントも削除
	batch.Delete(r.client.Collection(roomCollection).Doc(roomID))

	_, err = batch.Commit(ctx)
	return err
}

// PlayerRepository は Firestore を使った PlayerRepository の実装
type PlayerRepository struct {
	client *firestore.Client
}

// NewPlayerRepository は PlayerRepository を作成する
func NewPlayerRepository(client *firestore.Client) repository.PlayerRepository {
	return &PlayerRepository{
		client: client,
	}
}

// FindByID は指定されたIDのプレイヤーを取得する
func (r *PlayerRepository) FindByID(ctx context.Context, roomID, userID string) (*entity.Player, error) {
	doc, err := r.client.Collection(roomCollection).Doc(roomID).
		Collection(playerSubCollection).Doc(userID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var player entity.Player
	if err := doc.DataTo(&player); err != nil {
		return nil, err
	}
	return &player, nil
}

// FindAllByRoomID は指定された部屋の全プレイヤーを取得する
func (r *PlayerRepository) FindAllByRoomID(ctx context.Context, roomID string) ([]*entity.Player, error) {
	docs, err := r.client.Collection(roomCollection).Doc(roomID).
		Collection(playerSubCollection).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	players := make([]*entity.Player, 0, len(docs))
	for _, doc := range docs {
		var player entity.Player
		if err := doc.DataTo(&player); err != nil {
			return nil, err
		}
		players = append(players, &player)
	}
	return players, nil
}

// Create はプレイヤーを作成する
func (r *PlayerRepository) Create(ctx context.Context, roomID, userID string, player *entity.Player) error {
	_, err := r.client.Collection(roomCollection).Doc(roomID).
		Collection(playerSubCollection).Doc(userID).Set(ctx, player)
	return err
}

// Update はプレイヤー情報を更新する
func (r *PlayerRepository) Update(ctx context.Context, roomID, userID string, player *entity.Player) error {
	_, err := r.client.Collection(roomCollection).Doc(roomID).
		Collection(playerSubCollection).Doc(userID).Set(ctx, player)
	return err
}

// Delete はプレイヤーを削除する
func (r *PlayerRepository) Delete(ctx context.Context, roomID, userID string) error {
	_, err := r.client.Collection(roomCollection).Doc(roomID).
		Collection(playerSubCollection).Doc(userID).Delete(ctx)
	return err
}

// CountByRoomID は指定された部屋のプレイヤー数を取得する
func (r *PlayerRepository) CountByRoomID(ctx context.Context, roomID string) (int, error) {
	docs, err := r.client.Collection(roomCollection).Doc(roomID).
		Collection(playerSubCollection).Documents(ctx).GetAll()
	if err != nil {
		return 0, err
	}
	return len(docs), nil
}

// ClearAllVotes は全プレイヤーの投票状態をリセットする
func (r *PlayerRepository) ClearAllVotes(ctx context.Context, roomID string) error {
	docs, err := r.client.Collection(roomCollection).Doc(roomID).
		Collection(playerSubCollection).Documents(ctx).GetAll()
	if err != nil {
		return err
	}

	// バッチで更新
	batch := r.client.Batch()
	for _, doc := range docs {
		batch.Update(doc.Ref, []firestore.Update{
			{Path: "currentVote", Value: ""},
		})
	}

	_, err = batch.Commit(ctx)
	return err
}
