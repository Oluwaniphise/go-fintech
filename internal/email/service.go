package email

import (
	"fmt"
	"net/smtp"
	"os"
)

type EmailService struct {
	From     string
	SMTPHost string
	SMTPPort string
	Username string
	Password string
}

func NewEmailService() *EmailService {
	return &EmailService{
		From:     os.Getenv("EMAIL_FROM"),
		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
	}
}

func (s *EmailService) SendVerificationEmail(to, verificationURL string) error {
	subject := "Verify your email address"

	body := fmt.Sprintf(`
Hello,

Please verify your email address by clicking the link below:

%s

This link will expire soon.

If you did not create this account, you can ignore this email.
`, verificationURL)

	message := []byte(
		"From: " + s.From + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=\"UTF-8\"\r\n" +
			"\r\n" +
			body,
	)

	auth := smtp.PlainAuth("", s.Username, s.Password, s.SMTPHost)

	return smtp.SendMail(
		s.SMTPHost+":"+s.SMTPPort,
		auth,
		s.From,
		[]string{to},
		message,
	)
}
