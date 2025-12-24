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

	verificationLink := fmt.Sprintf(`https://nedzl.com/auth/verify?token=%s&email=%s`, token, to)

	html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="utf-8">
        <style>
            body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f4f7f6; margin: 0; padding: 0; -webkit-font-smoothing: antialiased; }
            .container { max-width: 600px; margin: 20px auto; background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.08); }
            .header { background: #07B463; padding: 40px 20px; text-align: center; }
            .header h1 { color: #ffffff; margin: 0; font-size: 28px; font-weight: 700; letter-spacing: -0.5px; }
            .content { padding: 40px; color: #333333; line-height: 1.6; }
            .content h2 { color: #07B463; font-size: 20px; margin-top: 0; }
            .btn { display: inline-block; background: #07B463; color: #ffffff !important; padding: 14px 30px; border-radius: 8px; text-decoration: none; font-weight: 600; font-size: 16px; transition: background 0.3s ease; }
            .footer { background: #f9fafb; padding: 20px; text-align: center; color: #718096; font-size: 13px; border-top: 1px solid #edf2f7; }
            .social-links { margin-top: 10px; }
            .social-links a { color: #07B463; text-decoration: none; margin: 0 10px; font-weight: 600; }
        </style>
    </head>
    <body>
        <div class="container">
            <div class="header">
                <h1>NedZl</h1>
            </div>
            <div class="content">
                <h2>Welcome to NedZl, %s!</h2>
                <p>We're excited to have you join our marketplace. To get started, please verify your email address by clicking the button below:</p>
                
                <div style="text-align: center; margin: 35px 0;">
                    <a href="%s" class="btn">Verify My Email</a>
                </div>

                <p>If the button doesn't work, you can also copy and paste this link into your browser:</p>
                <p style="word-break: break-all; color: #07B463; font-size: 14px;">%s</p>

                <p style="margin-top: 30px;">If you didn't create an account with us, you can safely ignore this email.</p>
                <p>Best regards,<br>The NedZl Team</p>
            </div>
            <div class="footer">
                <p>&copy; 2025 NedZl Marketplace. All rights reserved.</p>
                <div class="social-links">
                    <a href="#">Help Center</a> | <a href="#">Terms of Service</a> | <a href="#">Privacy Policy</a>
                </div>
            </div>
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
	html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="utf-8">
        <style>
            body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f4f7f6; margin: 0; padding: 0; -webkit-font-smoothing: antialiased; }
            .container { max-width: 600px; margin: 20px auto; background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.08); }
            .header { background: #4A5568; padding: 40px 20px; text-align: center; }
            .header h1 { color: #ffffff; margin: 0; font-size: 28px; font-weight: 700; letter-spacing: -0.5px; }
            .content { padding: 40px; color: #333333; line-height: 1.6; }
            .content h2 { color: #4A5568; font-size: 20px; margin-top: 0; }
            .alert-card { background: #fff5f5; border: 1px solid #fed7d7; border-radius: 8px; padding: 20px; margin: 25px 0; border-left: 4px solid #F56565; }
            .btn { display: inline-block; background: #4A5568; color: #ffffff !important; padding: 14px 30px; border-radius: 8px; text-decoration: none; font-weight: 600; font-size: 16px; }
            .footer { background: #f9fafb; padding: 20px; text-align: center; color: #718096; font-size: 13px; border-top: 1px solid #edf2f7; }
            .social-links { margin-top: 10px; }
            .social-links a { color: #4A5568; text-decoration: none; margin: 0 10px; font-weight: 600; }
        </style>
    </head>
    <body>
        <div class="container">
            <div class="header">
                <h1>NedZl</h1>
            </div>
            <div class="content">
                <h2>Hello %s,</h2>
                
                <div class="alert-card">
                    <p style="margin: 0; font-weight: 600; color: #C53030;">Important Notice</p>
                    <p style="margin: 5px 0 0 0;">Your NedZl account has been <strong style="color: #C53030;">deactivated</strong>.</p>
                </div>

                <p>If you believe this has happened in error, or if you would like to appeal this decision, please reach out to our support team.</p>

                <div style="text-align: center; margin: 35px 0;">
                    <a href="mailto:support@nedzl.com" class="btn">Contact Support</a>
                </div>

                <p>Best regards,<br>The NedZl Team</p>
            </div>
            <div class="footer">
                <p>&copy; 2025 NedZl Marketplace. All rights reserved.</p>
                <div class="social-links">
                    <a href="#">Help Center</a> | <a href="#">Terms of Service</a> | <a href="#">Privacy Policy</a>
                </div>
            </div>
        </div>
    </body>
    </html>`, username)

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
	// product mail
	client := resend.NewClient(apiKey)
	html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="utf-8">
        <style>
            body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f4f7f6; margin: 0; padding: 0; -webkit-font-smoothing: antialiased; }
            .container { max-width: 600px; margin: 20px auto; background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.08); }
            .header { background: #4A5568; padding: 40px 20px; text-align: center; }
            .header h1 { color: #ffffff; margin: 0; font-size: 28px; font-weight: 700; letter-spacing: -0.5px; }
            .content { padding: 40px; color: #333333; line-height: 1.6; }
            .content h2 { color: #4A5568; font-size: 20px; margin-top: 0; }
            .product-card { background: #f9fafb; border: 1px solid #edf2f7; border-radius: 8px; padding: 20px; margin: 25px 0; border-left: 4px solid #4A5568; }
            .product-name { font-size: 18px; font-weight: 600; color: #2d3748; margin-bottom: 5px; }
            .reason-label { display: block; font-size: 12px; font-weight: 700; color: #718096; text-transform: uppercase; margin-bottom: 4px; }
            .reason-box { background: #fffaf0; border: 1px solid #fbd38d; border-radius: 6px; padding: 12px; margin-top: 10px; font-size: 14px; color: #7b341e; }
            .btn { display: inline-block; background: #4A5568; color: #ffffff !important; padding: 14px 30px; border-radius: 8px; text-decoration: none; font-weight: 600; font-size: 16px; }
            .footer { background: #f9fafb; padding: 20px; text-align: center; color: #718096; font-size: 13px; border-top: 1px solid #edf2f7; }
            .social-links { margin-top: 10px; }
            .social-links a { color: #4A5568; text-decoration: none; margin: 0 10px; font-weight: 600; }
        </style>
    </head>
    <body>
        <div class="container">
            <div class="header">
                <h1>NedZl</h1>
            </div>
            <div class="content">
                <h2>Hello %s,</h2>
                <p>We are writing to let you know that your product listing has been <strong style="color: #4A5568;">removed</strong> from the NedZl marketplace.</p>
                
                <div class="product-card">
                    <span class="reason-label">Product Removed</span>
                    <div class="product-name">%s</div>
                    <div class="reason-box">
                        <strong>Reason provided:</strong><br>
                        %s
                    </div>
                </div>

                <p>If you think this is a mistake, or if you'd like to understand more about our listing policies, please feel free to contact our support team.</p>

                <div style="text-align: center; margin: 35px 0;">
                    <a href="mailto:support@nedzl.com" class="btn">Contact Support</a>
                </div>

                <p>Best regards,<br>The NedZl Team</p>
            </div>
            <div class="footer">
                <p>&copy; 2025 NedZl Marketplace. All rights reserved.</p>
                <div class="social-links">
                    <a href="#">Help Center</a> | <a href="#">Terms of Service</a> | <a href="#">Privacy Policy</a>
                </div>
            </div>
        </div>
    </body>
    </html>`, username, productname, reason)
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
	html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="utf-8">
        <style>
            body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f4f7f6; margin: 0; padding: 0; -webkit-font-smoothing: antialiased; }
            .container { max-width: 600px; margin: 20px auto; background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.08); }
            .header { background: #07B463; padding: 40px 20px; text-align: center; }
            .header h1 { color: #ffffff; margin: 0; font-size: 28px; font-weight: 700; letter-spacing: -0.5px; }
            .content { padding: 40px; color: #333333; line-height: 1.6; }
            .content h2 { color: #07B463; font-size: 20px; margin-top: 0; }
            .success-card { background: #e6fffa; border: 1px solid #b2f5ea; border-radius: 8px; padding: 20px; margin: 25px 0; border-left: 4px solid #07B463; }
            .btn { display: inline-block; background: #07B463; color: #ffffff !important; padding: 14px 30px; border-radius: 8px; text-decoration: none; font-weight: 600; font-size: 16px; }
            .footer { background: #f9fafb; padding: 20px; text-align: center; color: #718096; font-size: 13px; border-top: 1px solid #edf2f7; }
            .social-links { margin-top: 10px; }
            .social-links a { color: #07B463; text-decoration: none; margin: 0 10px; font-weight: 600; }
        </style>
    </head>
    <body>
        <div class="container">
            <div class="header">
                <h1>NedZl</h1>
            </div>
            <div class="content">
                <h2>Congratulations, %s!</h2>
                
                <div class="success-card">
                    <p style="margin: 0; font-weight: 600; color: #234e52;">Account Verified Successfully</p>
                    <p style="margin: 5px 0 0 0;">Your NedZl account has been fully verified. You can now start exploring the marketplace and listing your products.</p>
                </div>

                <p>Log in to your account now to get started with your marketplace journey.</p>

                <div style="text-align: center; margin: 35px 0;">
                    <a href="https://nedzl-market.vercel.app/login" class="btn">Login to My Account</a>
                </div>

                <p>Thank you for choosing NedZl!</p>
                <p>Best regards,<br>The NedZl Team</p>
            </div>
            <div class="footer">
                <p>&copy; 2025 NedZl Marketplace. All rights reserved.</p>
                <div class="social-links">
                    <a href="#">Help Center</a> | <a href="#">Terms of Service</a> | <a href="#">Privacy Policy</a>
                </div>
            </div>
        </div>
    </body>
    </html>`, username)

	params := &resend.SendEmailRequest{
		From:    "noreply@nedzl.com",
		To:      []string{to},
		Html:    html,
		Subject: "Account Verified Successfully",
	}

	_, err := client.Emails.Send(params)

	return err
}

func SendProductReactivationEmail(to, username, productname, productID string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	client := resend.NewClient(apiKey)

	productLink := fmt.Sprintf("https://nedzl.com/product/%s", productID)

	html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="utf-8">
        <style>
            body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f4f7f6; margin: 0; padding: 0; -webkit-font-smoothing: antialiased; }
            .container { max-width: 600px; margin: 20px auto; background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.08); }
            .header { background: #07B463; padding: 40px 20px; text-align: center; }
            .header h1 { color: #ffffff; margin: 0; font-size: 28px; font-weight: 700; letter-spacing: -0.5px; }
            .content { padding: 40px; color: #333333; line-height: 1.6; }
            .content h2 { color: #07B463; font-size: 20px; margin-top: 0; }
            .product-card { background: #f9fafb; border: 1px solid #edf2f7; border-radius: 8px; padding: 20px; margin: 25px 0; border-left: 4px solid #07B463; }
            .product-name { font-size: 18px; font-weight: 600; color: #2d3748; margin-bottom: 5px; }
            .status-badge { display: inline-block; background: #e6fffa; color: #047481; font-size: 12px; font-weight: 700; padding: 4px 12px; border-radius: 20px; text-transform: uppercase; margin-bottom: 10px; }
            .btn { display: inline-block; background: #07B463; color: #ffffff !important; padding: 14px 30px; border-radius: 8px; text-decoration: none; font-weight: 600; font-size: 16px; transition: background 0.3s ease; }
            .footer { background: #f9fafb; padding: 20px; text-align: center; color: #718096; font-size: 13px; border-top: 1px solid #edf2f7; }
            .social-links { margin-top: 10px; }
            .social-links a { color: #07B463; text-decoration: none; margin: 0 10px; font-weight: 600; }
        </style>
    </head>
    <body>
        <div class="container">
            <div class="header">
                <h1>NedZl</h1>
            </div>
            <div class="content">
                <h2>Good news, %s!</h2>
                <p>Your product has been successfully <strong style="color: #07B463;">reactivated</strong> and is now visible to buyers across our marketplace.</p>
                
                <div class="product-card">
                    <span class="status-badge">Live Now</span>
                    <div class="product-name">%s</div>
                    <p style="margin: 0; font-size: 14px; color: #718096;">It's time to start receiving offers again. Make sure your details are up to date to close the deal faster!</p>
                </div>

                <div style="text-align: center; margin-top: 35px;">
                    <a href="%s" class="btn">View Product Listing</a>
                </div>

                <p style="margin-top: 30px;">If you have any questions, our support team is always here to help.</p>
                <p>Best regards,<br>The NedZl Team</p>
            </div>
            <div class="footer">
                <p>&copy; 2025 NedZl Marketplace. All rights reserved.</p>
                <div class="social-links">
                    <a href="#">Help Center</a> | <a href="#">Terms of Service</a> | <a href="#">Privacy Policy</a>
                </div>
            </div>
        </div>
    </body>
    </html>`, username, productname, productLink)

	params := &resend.SendEmailRequest{
		From:    "noreply@nedzl.com",
		To:      []string{to},
		Html:    html,
		Subject: fmt.Sprintf("Success! Your product '%s' is back online", productname),
	}

	_, err := client.Emails.Send(params)
	return err
}

func SendProductClosureEmail(to, username, productname string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	client := resend.NewClient(apiKey)

	html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="utf-8">
        <style>
            body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f4f7f6; margin: 0; padding: 0; -webkit-font-smoothing: antialiased; }
            .container { max-width: 600px; margin: 20px auto; background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.08); }
            .header { background: #4A5568; padding: 40px 20px; text-align: center; }
            .header h1 { color: #ffffff; margin: 0; font-size: 28px; font-weight: 700; letter-spacing: -0.5px; }
            .content { padding: 40px; color: #333333; line-height: 1.6; }
            .content h2 { color: #4A5568; font-size: 20px; margin-top: 0; }
            .info-card { background: #f7fafc; border: 1px solid #edf2f7; border-radius: 8px; padding: 20px; margin: 25px 0; border-left: 4px solid #4A5568; }
            .product-name { font-size: 18px; font-weight: 600; color: #2d3748; margin-bottom: 5px; }
            .status-badge { display: inline-block; background: #edf2f7; color: #4a5568; font-size: 12px; font-weight: 700; padding: 4px 12px; border-radius: 20px; text-transform: uppercase; margin-bottom: 10px; }
            .footer { background: #f9fafb; padding: 20px; text-align: center; color: #718096; font-size: 13px; border-top: 1px solid #edf2f7; }
            .social-links { margin-top: 10px; }
            .social-links a { color: #4A5568; text-decoration: none; margin: 0 10px; font-weight: 600; }
        </style>
    </head>
    <body>
        <div class="container">
            <div class="header">
                <h1>NedZl</h1>
            </div>
            <div class="content">
                <h2>Hello %s,</h2>
                <p>We are writing to inform you that your product listing has been <strong style="color: #4A5568;">closed</strong>.</p>
                
                <div class="info-card">
                    <span class="status-badge">Listing Closed</span>
                    <div class="product-name">%s</div>
                    <p style="margin: 0; font-size: 14px; color: #718096;">If you sold this product, congratulations! If you'd like to relist it or have questions about why it was closed, feel free to visit your dashboard or contact our support.</p>
                </div>

                <div style="text-align: center; margin-top: 35px;">
                    <a href="https://nedzl.com/dashboard/products" style="display: inline-block; background: #4A5568; color: #ffffff !important; padding: 14px 30px; border-radius: 8px; text-decoration: none; font-weight: 600; font-size: 16px;">Go to Dashboard</a>
                </div>

                <p style="margin-top: 30px;">Thank you for using NedZl Marketplace.</p>
                <p>Best regards,<br>The NedZl Team</p>
            </div>
            <div class="footer">
                <p>&copy; 2025 NedZl Marketplace. All rights reserved.</p>
                <div class="social-links">
                    <a href="#">Help Center</a> | <a href="#">Terms of Service</a> | <a href="#">Privacy Policy</a>
                </div>
            </div>
        </div>
    </body>
    </html>`, username, productname)

	params := &resend.SendEmailRequest{
		From:    "noreply@nedzl.com",
		To:      []string{to},
		Html:    html,
		Subject: fmt.Sprintf("Notice: Your product listing '%s' has been closed", productname),
	}

	_, err := client.Emails.Send(params)
	return err
}
