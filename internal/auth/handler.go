package auth

import (
	"errors"
	"fintech/internal/common"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
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

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phoneNumber"`
	} `json:"user"`
}

type VerifyEmailRequest struct {
	Token string `json:"token"`
}

type ResendVerificationEmailRequest struct {
	Email string `json:"email"`
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			message := "User already exists"
			details := pgErr.Message

			switch pgErr.ConstraintName {
			case "idx_users_email":
				message = "User with email already exists"
			case "idx_users_phone_number":
				message = "User with phone number already exists"
			}

			return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
				fiber.StatusInternalServerError,
				"AUTH_REGISTER_FAILED",
				message,
				common.ErrorDetail{Details: details},
			))
		}
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
		"User created successfully. Please check your email to verify your account.",
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

	loginResult, err := s.Login(req.Email, req.Password)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(common.Failure(
			fiber.StatusUnauthorized,
			"AUTH_INVALID_CREDENTIALS",
			"Invalid email or password",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	response := LoginResponse{
		Token: loginResult.Token,
	}
	response.User.FirstName = loginResult.User.FirstName
	response.User.LastName = loginResult.User.LastName
	response.User.Email = loginResult.User.Email
	response.User.PhoneNumber = loginResult.User.PhoneNumber

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		"AUTH_LOGIN_SUCCESS",
		"Login successful",
		response,
	))
}

func (s *AuthService) HandleVerifyEmail(c *fiber.Ctx) error {
	req := new(VerifyEmailRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"AUTH_INVALID_REQUEST",
			"Invalid request",
			common.ErrorDetail{Details: "request body could not be parsed"},
		))
	}

	if err := s.VerifyEmail(req.Token); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"AUTH_EMAIL_VERIFICATION_FAILED",
			"Invalid or expired verification token",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		"AUTH_EMAIL_VERIFIED",
		"Email verified successfully",
		struct{}{},
	))
}

func (s *AuthService) HandleResendVerificationEmail(c *fiber.Ctx) error {
	req := new(ResendVerificationEmailRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"AUTH_INVALID_REQUEST",
			"Invalid request",
			common.ErrorDetail{Details: "request body could not be parsed"},
		))
	}

	if err := s.ResendVerificationEmail(req.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"AUTH_RESEND_VERIFICATION_FAILED",
			"Could not resend verification email",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		"AUTH_VERIFICATION_EMAIL_SENT",
		"Verification email sent successfully",
		struct{}{},
	))
}
