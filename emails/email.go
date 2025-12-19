package emails

import (
	"fmt"
	"os"
	"sync"

	"github.com/resend/resend-go/v3"
)

var (
	Client *resend.Client
	once   sync.Once
)

func InitEmailClient() {
	once.Do(func() {
		apiKey := os.Getenv("RESEND_API_KEY")
		if apiKey == "" {
			fmt.Println("❌ RESEND_API_KEY is EMPTY in InitEmailClient")
		} else {
			fmt.Printf("✅ RESEND_API_KEY loaded: %s...\n", apiKey[:10]) // Print first 10 chars
		}
		Client = resend.NewClient(apiKey)
	})
}

func SendVerificationMail(to, username, token string) error {
	if Client == nil {
		return fmt.Errorf("email client not initialized")
	}

	verificationLink := fmt.Sprintf(`https://nedzl-market.vercel.app/auth/verify?token=%s&email=%s`, token, to)

	html := fmt.Sprintf(`<html>
        <body style="background:#F5F5F5;padding:40px;font-family:Arial;">
            <div style="max-width:600px;margin:auto;background:#fff;padding:30px;border-radius:10px;">
                <h2 style="color:#07B463;text-align:center;">NedZl</h2>
                <p>Hello %s,</p>
                <p>Click the button below to verify your email address:</p>

                <a href="%s" 
                   style="background:#07B463;color:white;padding:14px 28px;border-radius:6px;text-decoration:none;">
                    Verify Email
                </a>

                <p>If button does not work, copy link below:</p>
                <p style="word-break:break-all;color:#07B463;">%s</p>
            </div>
        </body>
        </html>`, username, verificationLink, verificationLink)

	params := &resend.SendEmailRequest{
		From:    "noreply@nedzl.com",
		To:      []string{to},
		Html:    html,
		Subject: "Verify your NedZl email",
	}

	fmt.Printf("Sending verification email to %s\n", to)
	_, err := Client.Emails.Send(params)

	return err
}

func SendUserDeactivationEmail(to, username string) error {
	apiKey := os.Getenv("RESEND_API_KEY")

	client := resend.NewClient(apiKey)
	html := fmt.Sprintf(`   <html><body style="background:#F5F5F5;padding:40px;font-family:Arial;">
            <div style="max-width:600px;margin:auto;background:#fff;padding:30px;border-radius:10px;">
                <h2 style="color:#07B463;text-align:center;">NedZl</h2>
                <p>Hello %s,</p>
                <p>Your NedZl account has been <strong style="color:#07B463;">deactivated</strong>.</p>
                <p>If this is an error, contact support team.</p>

                <a href="mailto:support@nedzl.com"
                   style="background:#07B463;color:white;padding:14px 28px;border-radius:6px;text-decoration:none;">
                    Contact Support
                </a>
            </div>
        </body></html>`, username)

	params := &resend.SendEmailRequest{
		From:    "noreply@nedzl.com",
		To:      []string{to},
		Html:    html,
		Subject: "Your NedZl Account Was Deactivated",
	}

	_, err := client.Emails.Send(params)

	return err
}

func SendProductDeactivationEmail(to, username, productname, reason string) error {
	apiKey := os.Getenv("RESEND_API_KEY")

	client := resend.NewClient(apiKey)
	html := fmt.Sprintf(`<html><body style="background:#F5F5F5;padding:40px;font-family:Arial;">
            <div style="max-width:600px;margin:auto;background:#fff;padding:30px;border-radius:10px;">
                <h2 style="color:#07B463;text-align:center;">NedZl</h2>
                <p>Hello %s,</p>
                <p>Your product <strong>"%s"</strong> has been removed from NedZl.</p>
                <p>Reason: %s</p>
                <p>If you think this is an error, contact support.</p>

                <a href="mailto:support@nedzl.com"
                   style="background:#07B463;color:white;padding:14px 28px;border-radius:6px;text-decoration:none;">
                    Contact Support
                </a>
            </div>
        </body></html>`, username, productname, reason)
	params := &resend.SendEmailRequest{
		From:    "noreply@nedzl.com",
		To:      []string{to},
		Html:    html,
		Subject: "Your Product Was Removed",
	}

	_, err := client.Emails.Send(params)

	return err
}

func SendAccountVerifiedMail(to, username string) error {
	apiKey := os.Getenv("RESEND_API_KEY")

	client := resend.NewClient(apiKey)
	html := fmt.Sprintf(`<html><body style="background:#F5F5F5;padding:40px;font-family:Arial;">
            <div style="max-width:600px;margin:auto;background:#fff;padding:30px;border-radius:10px;">
                <h2 style="color:#07B463;text-align:center;">NedZl</h2>
                <p>Hello %s,</p>
                <p>Your account has been successfully verified! You can now log in and start using NedZl.</p>
                
                <a href="https://nedzl-market.vercel.app/login" 
                   style="background:#07B463;color:white;padding:14px 28px;border-radius:6px;text-decoration:none;display:inline-block;margin-top:20px;">
                    Login to Account
                </a>
            </div>
        </body></html>`, username)

	params := &resend.SendEmailRequest{
		From:    "noreply@nedzl.com",
		To:      []string{to},
		Html:    html,
		Subject: "Account Verified Successfully",
	}

	_, err := client.Emails.Send(params)

	return err
}
