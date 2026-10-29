package zatca

import "github.com/NasTecSol/nembus-core/usecase"

type environmentClientFactory func(string) (*ZatcaClient, error)

func newEnvironmentClient(environment string) (*ZatcaClient, error) {
	baseURL, err := usecase.ZatcaEnvironmentBaseURL(environment)
	if err != nil {
		return nil, err
	}
	return NewZatcaClient(baseURL), nil
}
