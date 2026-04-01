package auth

import "github.com/gofiber/fiber/v2"

type RegisterRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

func (s *AuthService) HandleRegister(c *fiber.Ctx) error {
	req := new(RegisterRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	user, err := s.RegisterUser(req.FirstName, req.LastName, req.Email, req.PhoneNumber, req.Password)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create user"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "User created successfully",
		"user_id": user.ID,
	})
}
