package usecase

import (
	"context"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/domain/repository"
)

// findPolicy は政策を取得する（AI生成 → マスターの順で探す）
func findPolicy(ctx context.Context, room *entity.Room, policyRepo repository.PolicyRepository, policyID string) (*entity.MasterPolicy, error) {
	// まずAI生成政策から探す
	policy := room.GetGeneratedPolicy(policyID)
	if policy != nil {
		return policy, nil
	}
	// なければマスターから探す
	return policyRepo.FindByID(ctx, policyID)
}
