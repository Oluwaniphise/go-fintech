package email

import (
	"fmt"
	"os"

	"github.com/resend/resend-go/v3"
)

type EmailService struct {
	Client *resend.Client
	From   string
}

func NewEmailService() *EmailService {
	apiKey := os.Getenv("RESEND_API_KEY")

	return &EmailService{
		Client: resend.NewClient(apiKey),
		From:   os.Getenv("EMAIL_FROM"),
	}
}

func (s *EmailService) SendVerificationEmail(to, verificationURL string) error {
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "Go Fintech"
	}

	params := &resend.SendEmailRequest{
		From:    s.From,
		To:      []string{to},
		Subject: "Verify your " + appName + " account",
		Html:    verificationEmailHTML(appName, verificationURL),
	}

	_, err := s.Client.Emails.Send(params)
	return err
}

func verificationEmailHTML(appName, verificationURL string) string {
	return fmt.Sprintf(`<!doctype html>
<html>
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Verify your email</title>
  </head>
  <body style="margin:0; padding:0; background-color:#f4f7fb; font-family:Arial, Helvetica, sans-serif;">
    <table width="100%%" cellpadding="0" cellspacing="0" role="presentation" style="background-color:#f4f7fb; padding:40px 16px;">
      <tr>
        <td align="center">
          <table width="100%%" cellpadding="0" cellspacing="0" role="presentation" style="max-width:560px; background-color:#ffffff; border-radius:18px; overflow:hidden; box-shadow:0 16px 40px rgba(15, 23, 42, 0.10);">
            <tr>
              <td style="background:linear-gradient(135deg,#0f766e,#14b8a6); padding:34px 30px; text-align:center;">
                <p style="margin:0 0 10px; color:#ccfbf1; font-size:13px; letter-spacing:0.12em; text-transform:uppercase; font-weight:700;">%s</p>
                <h1 style="margin:0; color:#ffffff; font-size:28px; line-height:1.2; font-weight:800;">Verify your email</h1>
                <p style="margin:12px 0 0; color:#ecfeff; font-size:15px; line-height:1.6;">One quick step to secure your account.</p>
              </td>
            </tr>

            <tr>
              <td style="padding:36px 32px 10px;">
                <h2 style="margin:0 0 14px; color:#0f172a; font-size:22px; line-height:1.3;">Confirm your email address</h2>
                <p style="margin:0 0 20px; color:#475569; font-size:15px; line-height:1.7;">
                  Thanks for creating an account with %s. Please verify your email address so we can keep your account protected.
                </p>

                <div style="text-align:center; margin:32px 0;">
                  <a href="%s" style="display:inline-block; background-color:#0f766e; color:#ffffff; text-decoration:none; padding:15px 30px; border-radius:999px; font-size:15px; font-weight:700;">
                    Verify Email Address
                  </a>
                </div>

                <p style="margin:0 0 18px; color:#64748b; font-size:14px; line-height:1.6;">
                  This verification link expires in 30 minutes. If you did not create this account, you can safely ignore this email.
                </p>

                <p style="margin:0; color:#94a3b8; font-size:13px; line-height:1.6;">
                  If the button does not work, copy and paste this link into your browser:
                </p>
                <p style="word-break:break-all; margin:8px 0 0; color:#0f766e; font-size:13px; line-height:1.6;">%s</p>
              </td>
            </tr>

            <tr>
              <td style="padding:24px 32px 32px;">
                <div style="height:1px; background-color:#e2e8f0; margin-bottom:20px;"></div>
                <p style="margin:0; text-align:center; color:#94a3b8; font-size:12px; line-height:1.6;">
                  You received this email because an account was created using this email address.
                </p>
                <p style="margin:8px 0 0; text-align:center; color:#94a3b8; font-size:12px;">
                  &copy; 2026 %s. All rights reserved.
                </p>
              </td>
            </tr>
          </table>
        </td>
      </tr>
    </table>
  </body>
</html>`, appName, appName, verificationURL, verificationURL, appName)
}
