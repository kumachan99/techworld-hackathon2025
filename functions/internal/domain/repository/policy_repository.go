package repository

import (
	"context"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
)

// PolicyRepository は政策マスターの永続化を担当するインターフェース
// パス: master_policies/{policyId}
type PolicyRepository interface {
	// GetAll は全ての政策マスターを取得する
	GetAll(ctx context.Context) ([]entity.MasterPolicy, error)

	// FindByID は指定されたIDの政策マスターを取得する
	FindByID(ctx context.Context, id string) (*entity.MasterPolicy, error)

	// FindByIDs は指定されたIDリストの政策マスターを取得する
	FindByIDs(ctx context.Context, ids []string) ([]entity.MasterPolicy, error)

	// GetAllIDs は全ての政策IDを取得する（デッキ作成用）
	GetAllIDs(ctx context.Context) ([]string, error)

	// Create は政策を作成する（AI陳情で生成された政策用）
	Create(ctx context.Context, policy *entity.MasterPolicy) (string, error)
}
