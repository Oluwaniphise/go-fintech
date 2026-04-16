package electricity

import "fintech/internal/bills"

type Service struct {
	Deps bills.Dependencies
}

func NewService(deps bills.Dependencies) *Service {
	return &Service{Deps: deps}
}
