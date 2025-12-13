package repository

import (
	"context"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
)

// IdeologyRepository は思想マスターの永続化を担当するインターフェース
// パス: master_ideologies/{ideologyId}
type IdeologyRepository interface {
	// GetAll は全ての思想マスターを取得する
	GetAll(ctx context.Context) ([]entity.MasterIdeology, error)

	// FindByID は指定されたIDの思想マスターを取得する
	FindByID(ctx context.Context, id string) (*entity.MasterIdeology, error)

	// GetAllIDs は全ての思想IDを取得する
	GetAllIDs(ctx context.Context) ([]string, error)
}
