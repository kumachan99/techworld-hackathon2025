package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/domain/repository"
)

const masterIdeologyCollection = "master_ideologies"

// IdeologyRepository は Firestore を使った IdeologyRepository の実装
type IdeologyRepository struct {
	client *firestore.Client
}

// NewIdeologyRepository は IdeologyRepository を作成する
func NewIdeologyRepository(client *firestore.Client) repository.IdeologyRepository {
	return &IdeologyRepository{
		client: client,
	}
}

// GetAll は全ての思想マスターを取得する
func (r *IdeologyRepository) GetAll(ctx context.Context) ([]entity.MasterIdeology, error) {
	docs, err := r.client.Collection(masterIdeologyCollection).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	ideologies := make([]entity.MasterIdeology, 0, len(docs))
	for _, doc := range docs {
		var ideology entity.MasterIdeology
		if err := doc.DataTo(&ideology); err != nil {
			return nil, err
		}
		ideology.IdeologyID = doc.Ref.ID
		ideologies = append(ideologies, ideology)
	}

	return ideologies, nil
}

// FindByID は指定されたIDの思想マスターを取得する
func (r *IdeologyRepository) FindByID(ctx context.Context, id string) (*entity.MasterIdeology, error) {
	doc, err := r.client.Collection(masterIdeologyCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var ideology entity.MasterIdeology
	if err := doc.DataTo(&ideology); err != nil {
		return nil, err
	}
	ideology.IdeologyID = doc.Ref.ID
	return &ideology, nil
}

// GetAllIDs は全ての思想IDを取得する
func (r *IdeologyRepository) GetAllIDs(ctx context.Context) ([]string, error) {
	docs, err := r.client.Collection(masterIdeologyCollection).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(docs))
	for _, doc := range docs {
		ids = append(ids, doc.Ref.ID)
	}

	return ids, nil
}
