package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
)

// Config holds email configuration
type Config struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromAddress  string
	FromName     string
}

// Mailer handles sending emails
type Mailer struct {
	config Config
}

// NewMailer creates a new mailer instance
func NewMailer(config Config) *Mailer {
	return &Mailer{config: config}
}

// WelcomeEmailData holds data for welcome email template
type WelcomeEmailData struct {
	Username string
	Email    string
	Language string
	LoginURL string
}

// SendWelcomeEmail sends a welcome email to a newly registered user
func (m *Mailer) SendWelcomeEmail(to string, data WelcomeEmailData) error {
	// Determine template based on language
	templateName := fmt.Sprintf("welcome_%s.html", data.Language)
	templatePath := filepath.Join("pkg", "email", "templates", templateName)

	// Fallback to English if language template doesn't exist
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		data.Language = "en"
		templatePath = filepath.Join("pkg", "email", "templates", "welcome_en.html")
	}

	// Read template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	// Execute template
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	// Get subject based on language
	subject := m.getWelcomeSubject(data.Language)

	// Send email
	return m.sendEmail(to, subject, body.String())
}

// getWelcomeSubject returns the subject line for welcome email based on language
func (m *Mailer) getWelcomeSubject(language string) string {
	subjects := map[string]string{
		"en": "Welcome to Recontext!",
		"ru": "Добро пожаловать в Recontext!",
	}

	subject, ok := subjects[language]
	if !ok {
		subject = subjects["en"] // Default to English
	}

	return subject
}

// sendEmail sends an email using SMTP
func (m *Mailer) sendEmail(to, subject, htmlBody string) error {
	// Build email message
	from := fmt.Sprintf("%s <%s>", m.config.FromName, m.config.FromAddress)

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%s", m.config.SMTPHost, m.config.SMTPPort)

	// Setup authentication
	auth := smtp.PlainAuth("", m.config.SMTPUser, m.config.SMTPPassword, m.config.SMTPHost)

	// Send email
	err := smtp.SendMail(addr, auth, m.config.FromAddress, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// LoadConfigFromEnv loads email configuration from environment variables
func LoadConfigFromEnv() Config {
	return Config{
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromAddress:  getEnv("SMTP_FROM_ADDRESS", "noreply@recontext.online"),
		FromName:     getEnv("SMTP_FROM_NAME", "Recontext"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}
