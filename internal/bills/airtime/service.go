package airtime

import "fintech/internal/bills"

type Service struct {
	Deps    bills.Dependencies
	Helpers *bills.Helpers
}

var AIRTIME_ENDPOINT = "vas/pay/airtime"

func NewService(deps bills.Dependencies) *Service {
	return &Service{
		Deps:    deps,
		Helpers: bills.NewHelpers(deps),
	}
}
