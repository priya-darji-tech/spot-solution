package email

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"gopkg.in/mail.v2"
	"spot-solution/internal/config"
	"spot-solution/internal/models"
	"spot-solution/internal/repository"
)

type Mailer struct {
	cfg  *config.Config
	dial *mail.Dialer
}

func NewMailer(cfg *config.Config) *Mailer {
	var dial *mail.Dialer
	if cfg.SMTPHost != "" && cfg.SMTPHost != "simulation" {
		dial = mail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword)
		if cfg.SMTPPort == 465 {
			dial.SSL = true
		}
	}
	return &Mailer{
		cfg:  cfg,
		dial: dial,
	}
}

func (m *Mailer) Send(to, subject, bodyHTML string) error {
	if m.dial == nil {
		// Log locally if no remote SMTP is configured
		log.Printf("[EMAIL SIMULATION] To: %s | Subject: %s\nBody: %s\n", to, subject, bodyHTML)
		return nil
	}

	msg := mail.NewMessage()
	msg.SetHeader("From", fmt.Sprintf("%s <%s>", m.cfg.SMTPFromName, m.cfg.SMTPFromEmail))
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", bodyHTML)

	if err := m.dial.DialAndSend(msg); err != nil {
		return fmt.Errorf("smtp delivery error: %w", err)
	}
	return nil
}

func (m *Mailer) BuildAdminNotificationHTML(sub *models.ContactSubmission) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>New Website Inquiry</title></head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="background: #0B132B; padding: 20px; border-radius: 8px 8px 0 0; text-align: center;">
    <h2 style="color: #00E5FF; margin: 0;">New Website Inquiry Received</h2>
  </div>
  <div style="border: 1px solid #e0e0e0; border-top: none; padding: 25px; border-radius: 0 0 8px 8px;">
    <p>A new potential customer submitted an inquiry via the website contact form:</p>
    <table style="width: 100%%; border-collapse: collapse;">
      <tr><td style="padding: 8px; font-weight: bold; width: 140px;">Name:</td><td style="padding: 8px;">%s %s</td></tr>
      <tr style="background: #f9f9f9;"><td style="padding: 8px; font-weight: bold;">Email:</td><td style="padding: 8px;"><a href="mailto:%s">%s</a></td></tr>
      <tr><td style="padding: 8px; font-weight: bold;">Phone:</td><td style="padding: 8px;">%s</td></tr>
      <tr style="background: #f9f9f9;"><td style="padding: 8px; font-weight: bold;">Company:</td><td style="padding: 8px;">%s</td></tr>
      <tr><td style="padding: 8px; font-weight: bold;">Service:</td><td style="padding: 8px; color: #0284c7; font-weight: bold;">%s</td></tr>
      <tr style="background: #f9f9f9;"><td style="padding: 8px; font-weight: bold;">Submitted:</td><td style="padding: 8px;">%s</td></tr>
    </table>
    <h4 style="margin-top: 20px; margin-bottom: 8px;">Message:</h4>
    <div style="background: #f1f5f9; padding: 15px; border-radius: 6px; border-left: 4px solid #00E5FF; white-space: pre-wrap;">%s</div>
  </div>
</body>
</html>
`, sub.FirstName, sub.LastName, sub.Email, sub.Email, sub.Phone, sub.CompanyName, sub.Service, sub.CreatedAt.Format("Jan 02, 2006 15:04:05 MST"), strings.ReplaceAll(sub.Message, "<", "&lt;"))
}

func (m *Mailer) BuildCustomerAcknowledgmentHTML(sub *models.ContactSubmission) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>Thank you for contacting %s</title></head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="background: #0B132B; padding: 25px; border-radius: 8px 8px 0 0; text-align: center;">
    <h2 style="color: #00E5FF; margin: 0;">%s</h2>
    <p style="color: #94a3b8; margin: 5px 0 0 0; font-size: 14px;">Intelligent Engineering. Scalable Solutions.</p>
  </div>
  <div style="border: 1px solid #e0e0e0; border-top: none; padding: 25px; border-radius: 0 0 8px 8px;">
    <p>Dear <strong>%s</strong>,</p>
    <p>Thank you for reaching out to us regarding <strong>%s</strong>. We have received your inquiry and our team is already reviewing your requirements.</p>
    <p>One of our senior technology consultants will get in touch with you within <strong>1 business day</strong> to discuss your project in detail.</p>
    <div style="background: #f8fafc; padding: 15px; border-radius: 6px; margin: 20px 0; border: 1px solid #e2e8f0;">
      <h4 style="margin: 0 0 8px 0; color: #1e293b;">Inquiry Summary:</h4>
      <p style="margin: 0; font-size: 14px; color: #64748b;"><strong>Reference ID:</strong> %s</p>
      <p style="margin: 5px 0 0 0; font-size: 14px; color: #64748b;"><strong>Service:</strong> %s</p>
    </div>
    <p>Best regards,<br><strong>The %s Team</strong></p>
  </div>
</body>
</html>
`, m.cfg.AppName, m.cfg.AppName, sub.FirstName, sub.Service, sub.ID.String()[:8], sub.Service, m.cfg.AppName)
}

// StartOutboxWorker runs a background worker loop processing pending emails with exponential backoff
func (m *Mailer) StartOutboxWorker(ctx context.Context, repo *repository.OutboxRepository) {
	interval := time.Duration(m.cfg.OutboxWorkerIntervalSec) * time.Second
	if interval < 2*time.Second {
		interval = 5 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting email outbox worker (polling every %v)...", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping email outbox worker...")
			return
		case <-ticker.C:
			items, err := repo.FetchPendingBatch(ctx, 10)
			if err != nil {
				log.Printf("Outbox worker query error: %v", err)
				continue
			}

			for _, item := range items {
				err := m.Send(item.RecipientEmail, item.Subject, item.BodyHTML)
				if err != nil {
					// Exponential backoff: 2^(retry_count) * 15 seconds
					backoffSec := math.Pow(2, float64(item.RetryCount)) * 15
					nextRetry := time.Now().Add(time.Duration(backoffSec) * time.Second)
					log.Printf("Failed to deliver outbox email ID %s to %s (attempt %d/%d): %v. Next retry in %.0fs",
						item.ID, item.RecipientEmail, item.RetryCount+1, item.MaxRetries, err, backoffSec)
					_ = repo.MarkFailed(ctx, item.ID, err.Error(), nextRetry)
				} else {
					log.Printf("Delivered outbox email ID %s to %s", item.ID, item.RecipientEmail)
					_ = repo.MarkSent(ctx, item.ID)
				}
			}
		}
	}
}
