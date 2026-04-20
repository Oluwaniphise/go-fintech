package auth

import (
	"errors"
	"fintech/internal/common"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
)

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

	result, err := s.StartLogin(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrEmailNotVerified) {
			return c.Status(fiber.StatusForbidden).JSON(common.Failure(
				fiber.StatusForbidden,
				"AUTH_EMAIL_NOT_VERIFIED",
				"Please verify your email before logging in.",
				nil,
			))
		}

		return c.Status(fiber.StatusUnauthorized).JSON(common.Failure(
			fiber.StatusUnauthorized,
			"AUTH_INVALID_CREDENTIALS",
			"Invalid email or password",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		"AUTH_OTP_SENT",
		"OTP sent successfully",
		LoginStartResponse{
			Reference: result.Reference,
		},
	))
}

func (s *AuthService) HandleVerifyLoginOTP(c *fiber.Ctx) error {
	req := new(VerifyLoginOTPRequest)
	if err := c.BodyParser(req); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	loginResult, err := s.VerifyLoginOTP(req.Reference, req.OTP)
	if err != nil {

		return c.Status(fiber.StatusUnauthorized).JSON(common.Failure(
			fiber.StatusUnauthorized,
			"AUTH_INVALID_OTP",
			"Invalid OTP",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	accessToken, err := s.SignAccessToken(loginResult.User)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.Failure(
			fiber.StatusInternalServerError,
			"AUTH_ACCESS_TOKEN_CREATION_FAILED",
			"Could not create access token",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		"AUTH_LOGIN_SUCCESSFUL",
		"login successful",
		struct {
			AccessToken string `json:"accessToken"`
			User        struct {
				ID          string `json:"id"`
				FirstName   string `json:"firstName"`
				LastName    string `json:"lastName"`
				Email       string `json:"email"`
				PhoneNumber string `json:"phoneNumber"`
			} `json:"user"`
		}{
			AccessToken: accessToken,
			User: struct {
				ID          string `json:"id"`
				FirstName   string `json:"firstName"`
				LastName    string `json:"lastName"`
				Email       string `json:"email"`
				PhoneNumber string `json:"phoneNumber"`
			}{
				ID:          loginResult.User.ID.String(),
				FirstName:   loginResult.User.FirstName,
				LastName:    loginResult.User.LastName,
				Email:       loginResult.User.Email,
				PhoneNumber: loginResult.User.PhoneNumber,
			},
		},
	))
}

func (s *AuthService) HandleRefresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies(refreshCookieName)
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing refresh token"})
	}

	session, err := s.RefreshSession(refreshToken, c.Get("User-Agent"), c.IP())
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid refresh token"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "session refreshed",
		"accessToken":  session.AccessToken,
		"refreshToken": session.RefreshToken,
	})
}

func (s *AuthService) HandleLogout(c *fiber.Ctx) error {
	refreshToken := c.Cookies(refreshCookieName)
	if refreshToken != "" {
		_ = s.RevokeSession(refreshToken)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "logged out",
	})
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

func (s *AuthService) HandleResendLoginOTP(c *fiber.Ctx) error {
	req := new(ResendLoginOTPRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"AUTH_INVALID_REQUEST",
			"Invalid request",
			common.ErrorDetail{Details: "request body could not be parsed"},
		))
	}

	result, err := s.ResendLoginOTP(req.Reference)
	if err != nil {
		if errors.Is(err, ErrOTPWaitFor1Minute) {
			return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
				fiber.StatusBadRequest,
				"AUTH_RESEND_OTP_FAILED",
				"Please wait for a minute before requesting for resend OTP.",
				common.ErrorDetail{Details: err.Error()},
			))
		}
		return c.Status(fiber.StatusBadRequest).JSON(common.Failure(
			fiber.StatusBadRequest,
			"AUTH_RESEND_OTP_FAILED",
			"Could not resend OTP",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		"AUTH_OTP_RESENT",
		"OTP resent successfully",
		ResendLoginOTPResponse{
			Reference: result.Reference,
		},
	))
}

func (s *AuthService) HandleLoggedInUserDetails(c *fiber.Ctx) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(common.Failure(
			fiber.StatusUnauthorized,
			"AUTH_UNAUTHORIZED",
			"Unauthorized",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	user, err := s.GetLoggedInUserDetails(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(common.Failure(
			fiber.StatusUnauthorized,
			"AUTH_USER_NOT_FOUND",
			"User not found",
			common.ErrorDetail{Details: err.Error()},
		))
	}

	return c.Status(fiber.StatusOK).JSON(common.Success(
		fiber.StatusOK,
		"AUTH_ME_SUCCESS",
		"Authenticated user fetched successfully",
		MeResponse{
			ID:          user.ID.String(),
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			IsVerified:  user.IsVerified,
		},
	))
}
