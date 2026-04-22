package betting

import (
	"fintech/internal/bills"

	"github.com/gofiber/fiber/v2"
)

var (
	VALIDATE_BETTING_ID = "vas/validate/betting"
	PAY_BETTING         = "vas/pay/betting"
)

func (s *Service) HandleBettingPurchase(c *fiber.Ctx) error {
	return bills.HandleBillPurchase(c, s.Deps, s.Helpers, bills.BillPurchaseConfig{
		Endpoint:       PAY_BETTING,
		Category:       "BETTING",
		Description:    "Betting wallet funding",
		FailureCode:    "BILL_BETTING_PURCHASE_FAILED",
		FailureMessage: "Betting purchase failed",
		SuccessCode:    "BILL_BETTING_PURCHASE_SUCCESS",
		SuccessMessage: "Betting purchase successful",
	})
}

func (s *Service) HandleValidateBetting(c *fiber.Ctx) error {
	return bills.HandleValidateBill(c, s.Deps, bills.BillPurchaseConfig{
		Endpoint:       VALIDATE_BETTING_ID,
		Category:       "BETTING",
		Description:    "Betting validated",
		FailureCode:    "BILL_BETTING_PURCHASE_FAILED",
		FailureMessage: "Could not validate betting ID",
		SuccessCode:    "BILL_BETTING_PURCHASE_SUCCESS",
		SuccessMessage: "Betting purchase successful",
	})
}
