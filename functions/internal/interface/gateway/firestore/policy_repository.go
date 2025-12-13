package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/domain/repository"
)

const masterPolicyCollection = "master_policies"

// PolicyRepository は Firestore を使った PolicyRepository の実装
type PolicyRepository struct {
	client *firestore.Client
}

// NewPolicyRepository は PolicyRepository を作成する
func NewPolicyRepository(client *firestore.Client) repository.PolicyRepository {
	return &PolicyRepository{
		client: client,
	}
}

// GetAll は全ての政策マスターを取得する
func (r *PolicyRepository) GetAll(ctx context.Context) ([]entity.MasterPolicy, error) {
	docs, err := r.client.Collection(masterPolicyCollection).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	policies := make([]entity.MasterPolicy, 0, len(docs))
	for _, doc := range docs {
		var policy entity.MasterPolicy
		if err := doc.DataTo(&policy); err != nil {
			return nil, err
		}
		policy.ID = doc.Ref.ID
		policies = append(policies, policy)
	}

	return policies, nil
}

// FindByID は指定されたIDの政策マスターを取得する
func (r *PolicyRepository) FindByID(ctx context.Context, id string) (*entity.MasterPolicy, error) {
	doc, err := r.client.Collection(masterPolicyCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var policy entity.MasterPolicy
	if err := doc.DataTo(&policy); err != nil {
		return nil, err
	}
	policy.ID = doc.Ref.ID
	return &policy, nil
}

// FindByIDs は指定されたIDリストの政策マスターを取得する
func (r *PolicyRepository) FindByIDs(ctx context.Context, ids []string) ([]entity.MasterPolicy, error) {
	policies := make([]entity.MasterPolicy, 0, len(ids))

	for _, id := range ids {
		policy, err := r.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if policy != nil {
			policies = append(policies, *policy)
		}
	}

	return policies, nil
}

// GetAllIDs は全ての政策IDを取得する（デッキ作成用）
func (r *PolicyRepository) GetAllIDs(ctx context.Context) ([]string, error) {
	docs, err := r.client.Collection(masterPolicyCollection).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(docs))
	for _, doc := range docs {
		ids = append(ids, doc.Ref.ID)
	}

	return ids, nil
}

// Create は政策を作成する（AI陳情で生成された政策用）
func (r *PolicyRepository) Create(ctx context.Context, policy *entity.MasterPolicy) (string, error) {
	docRef, _, err := r.client.Collection(masterPolicyCollection).Add(ctx, policy)
	if err != nil {
		return "", err
	}
	return docRef.ID, nil
}
