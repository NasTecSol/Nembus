package zatca

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/google/uuid"
)

func (s *Service) orderDocument(ctx context.Context, id uuid.UUID) (*repository.ZatcaDocumentChain, *repository.Queries, error) {
	repo, ok := ctx.Value(middleware.RepoKey).(*repository.Queries)
	if !ok || repo == nil {
		return nil, nil, fmt.Errorf("tenant repository is required for order ZATCA")
	}
	uc := usecase.NewZatcaUseCase(s.cfg)
	uc.SetRepository(repo)
	entry, err := uc.EnsureSalesOrderDocument(ctx, id)
	return entry, repo, err
}

func (s *Service) CheckSalesOrder(ctx context.Context, id uuid.UUID) (json.RawMessage, error) {
	entry, repo, err := s.orderDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.checkDocument(ctx, repo, entry)
}

func (s *Service) SubmitSalesOrder(ctx context.Context, id uuid.UUID) (*repository.ZatcaDocumentChain, error) {
	entry, repo, err := s.orderDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.submitDocument(ctx, repo, entry)
}
