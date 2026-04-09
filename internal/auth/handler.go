package auth

import (
	"fintech/internal/common"

	"github.com/gofiber/fiber/v2"
)

type RegisterRequest struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	Password    string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *AuthService) HandleRegister(c *fiber.Ctx) error {
	req := new(RegisterRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"AUTH_INVALID_REQUEST",
			"Invalid request",
			common.ErrorDetail{Details: "request body could not be parsed"},
		))
	}

	user, err := s.RegisterUser(req.FirstName, req.LastName, req.Email, req.PhoneNumber, req.Password)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"AUTH_REGISTER_FAILED",
			"Could not create user",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	return c.Status(fiber.StatusCreated).JSON(common.Success(
		fiber.StatusCreated,
		"AUTH_REGISTER_SUCCESS",
		"User created successfully",
		struct {
			UserID string `json:"userId"`
		}{
			UserID: user.ID.String(),
		},
	))
}

func (s *AuthService) HandleLogin(c *fiber.Ctx) error {
	req := new(LoginRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"AUTH_INVALID_REQUEST",
			"Invalid request",
			common.ErrorDetail{Details: "request body could not be parsed"},
		))
	}

	token, err := s.Login(req.Email, req.Password)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(common.Failure(
			fiber.StatusUnauthorized,
			"AUTH_INVALID_CREDENTIALS",
			"Invalid email or password",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		"AUTH_LOGIN_SUCCESS",
		"Login successful",
		struct {
			Token string `json:"token"`
		}{
			Token: token,
		},
	))
}
