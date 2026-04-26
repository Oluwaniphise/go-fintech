package betting

import "fintech/internal/bills"

type Service struct {
	Deps    bills.Dependencies
	Helpers *bills.Helpers
}

func NewService(deps bills.Dependencies) *Service {
	return &Service{
		Deps:    deps,
		Helpers: bills.NewHelpers(deps),
	}
}
