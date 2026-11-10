package notification

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

type SESConfig struct {
	Host      string `json:"host"`       // e.g. email-smtp.me-central-1.amazonaws.com
	Port      int    `json:"port"`       // e.g. 587
	Username  string `json:"username"`   // AWS SES SMTP Username
	Password  string `json:"password"`   // AWS SES SMTP Password
	FromEmail string `json:"from_email"` // Verified sender email address
	FromName  string `json:"from_name"`  // Sender display name (e.g. "Nembus ERP")
}

type SESClient struct {
	cfg SESConfig
}

func NewSESClient(cfg SESConfig) *SESClient {
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	if cfg.FromName == "" {
		cfg.FromName = "Nembus"
	}
	return &SESClient{cfg: cfg}
}

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

func (c *SESClient) SendEmail(ctx context.Context, toEmail, subject, htmlBody string, attachments ...Attachment) (string, error) {
	if c.cfg.Host == "" || c.cfg.Username == "" || c.cfg.Password == "" {
		return "", fmt.Errorf("aws ses credentials not configured")
	}

	fromHeader := fmt.Sprintf("%s <%s>", c.cfg.FromName, c.cfg.FromEmail)
	boundary := fmt.Sprintf("----=_Part_%d", time.Now().UnixNano())

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\n", fromHeader))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if len(attachments) > 0 {
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundary))
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		msg.WriteString(htmlBody + "\r\n\r\n")

		for _, att := range attachments {
			msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			msg.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", att.ContentType, att.Filename))
			msg.WriteString("Content-Transfer-Encoding: base64\r\n")
			msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", att.Filename))

			encoded := base64.StdEncoding.EncodeToString(att.Data)
			for i := 0; i < len(encoded); i += 76 {
				end := i + 76
				if end > len(encoded) {
					end = len(encoded)
				}
				msg.WriteString(encoded[i:end] + "\r\n")
			}
			msg.WriteString("\r\n")
		}
		msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		msg.WriteString(htmlBody)
	}

	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)
	auth := smtp.PlainAuth("", c.cfg.Username, c.cfg.Password, c.cfg.Host)

	// Connect to SMTP with STARTTLS
	client, err := smtp.Dial(addr)
	if err != nil {
		return "", fmt.Errorf("smtp dial failed: %w", err)
	}
	defer client.Close()

	tlsConfig := &tls.Config{
		ServerName: c.cfg.Host,
	}

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(tlsConfig); err != nil {
			return "", fmt.Errorf("starttls failed: %w", err)
		}
	}

	if err := client.Auth(auth); err != nil {
		return "", fmt.Errorf("smtp auth failed: %w", err)
	}

	if err := client.Mail(c.cfg.FromEmail); err != nil {
		return "", fmt.Errorf("mail from failed: %w", err)
	}

	if err := client.Rcpt(toEmail); err != nil {
		return "", fmt.Errorf("rcpt to failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return "", fmt.Errorf("data writer failed: %w", err)
	}

	if _, err := w.Write([]byte(msg.String())); err != nil {
		return "", fmt.Errorf("write email failed: %w", err)
	}

	if err := w.Close(); err != nil {
		return "", fmt.Errorf("close email data failed: %w", err)
	}

	_ = client.Quit()

	messageID := fmt.Sprintf("ses-%d-%s", time.Now().UnixNano(), toEmail)
	return messageID, nil
}
