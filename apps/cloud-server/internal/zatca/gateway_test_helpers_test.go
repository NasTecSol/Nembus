package zatca

import (
	"fmt"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
)

// Only tests inject local gateway URLs. Runtime factories always use the
// canonical allowlisted URL for the requested/saved device environment.
func testOnboardingService(repo *repository.Queries, client *ZatcaClient, cfg *usecase.ZatcaConfig) *OnboardingService {
	service := NewOnboardingService(repo, cfg)
	service.clients = func(env string) (*ZatcaClient, error) {
		if _, err := usecase.ZatcaEnvironmentBaseURL(env); err != nil {
			return nil, err
		}
		if client == nil {
			return nil, fmt.Errorf("test gateway not configured")
		}
		return client, nil
	}
	return service
}
func testService(repo *repository.Queries, cfg *usecase.ZatcaConfig) *Service {
	service := NewService(repo, cfg)
	service.clients = func(env string) (*ZatcaClient, error) {
		if _, err := usecase.ZatcaEnvironmentBaseURL(env); err != nil {
			return nil, err
		}
		if cfg.BaseURL == "" {
			return nil, fmt.Errorf("test gateway not configured")
		}
		return NewZatcaClient(cfg.BaseURL), nil
	}
	return service
}
