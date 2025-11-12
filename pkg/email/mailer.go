package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
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
	logger *log.Logger
}

// NewMailer creates a new mailer instance
func NewMailer(config Config) *Mailer {
	return &Mailer{
		config: config,
		logger: log.New(os.Stdout, "[EMAIL] ", log.LstdFlags),
	}
}

// WelcomeEmailData holds data for welcome email template
type WelcomeEmailData struct {
	Username string
	Email    string
	Password string
	Language string
	LoginURL string
}

// PasswordResetEmailData holds data for password reset email template
type PasswordResetEmailData struct {
	Email    string
	Code     string
	Language string
}

// SendWelcomeEmail sends a welcome email to a newly registered user
func (m *Mailer) SendWelcomeEmail(to string, data WelcomeEmailData) error {
	m.logger.Printf("Starting to send welcome email to %s (language: %s)", to, data.Language)

	// Determine template based on language
	templateName := fmt.Sprintf("welcome_%s.html", data.Language)
	templatePath := filepath.Join("pkg", "email", "templates", templateName)
	m.logger.Printf("Looking for template at: %s", templatePath)

	// Fallback to English if language template doesn't exist
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		m.logger.Printf("Template not found, falling back to English template")
		data.Language = "en"
		templatePath = filepath.Join("pkg", "email", "templates", "welcome_en.html")
	}

	// Read template
	m.logger.Printf("Parsing email template from: %s", templatePath)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		m.logger.Printf("ERROR: Failed to parse email template: %v", err)
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	// Execute template
	m.logger.Printf("Executing email template with data for user: %s", data.Username)
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		m.logger.Printf("ERROR: Failed to execute email template: %v", err)
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	// Get subject based on language
	subject := m.getWelcomeSubject(data.Language)
	m.logger.Printf("Email subject: %s", subject)

	// Send email
	m.logger.Printf("Attempting to send email to %s via SMTP", to)
	err = m.sendEmail(to, subject, body.String())
	if err != nil {
		m.logger.Printf("ERROR: Failed to send email: %v", err)
		return err
	}

	m.logger.Printf("SUCCESS: Welcome email sent successfully to %s", to)
	return nil
}

// SendPasswordResetEmail sends a password reset email with verification code
func (m *Mailer) SendPasswordResetEmail(to string, data PasswordResetEmailData) error {
	m.logger.Printf("Starting to send password reset email to %s (language: %s)", to, data.Language)

	// Determine template based on language
	templateName := fmt.Sprintf("password_reset_%s.html", data.Language)
	templatePath := filepath.Join("pkg", "email", "templates", templateName)
	m.logger.Printf("Looking for template at: %s", templatePath)

	// Fallback to English if language template doesn't exist
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		m.logger.Printf("Template not found, falling back to English template")
		data.Language = "en"
		templatePath = filepath.Join("pkg", "email", "templates", "password_reset_en.html")
	}

	// Read template
	m.logger.Printf("Parsing email template from: %s", templatePath)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		m.logger.Printf("ERROR: Failed to parse email template: %v", err)
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	// Execute template
	m.logger.Printf("Executing email template with reset code for: %s", data.Email)
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		m.logger.Printf("ERROR: Failed to execute email template: %v", err)
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	// Get subject based on language
	subject := m.getPasswordResetSubject(data.Language)
	m.logger.Printf("Email subject: %s", subject)

	// Send email
	m.logger.Printf("Attempting to send password reset email to %s via SMTP", to)
	err = m.sendEmail(to, subject, body.String())
	if err != nil {
		m.logger.Printf("ERROR: Failed to send email: %v", err)
		return err
	}

	m.logger.Printf("SUCCESS: Password reset email sent successfully to %s", to)
	return nil
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

// getPasswordResetSubject returns the subject line for password reset email based on language
func (m *Mailer) getPasswordResetSubject(language string) string {
	subjects := map[string]string{
		"en": "Password Reset - Recontext",
		"ru": "Сброс пароля - Recontext",
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
	m.logger.Printf("Building email from: %s, to: %s", from, to)

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// Log all email headers
	m.logger.Printf("=== EMAIL HEADERS ===")
	m.logger.Printf("From: %s", from)
	m.logger.Printf("To: %s", to)
	m.logger.Printf("Subject: %s", subject)
	m.logger.Printf("FromName config: '%s'", m.config.FromName)
	m.logger.Printf("FromAddress config: '%s'", m.config.FromAddress)
	m.logger.Printf("=====================")

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%s", m.config.SMTPHost, m.config.SMTPPort)
	m.logger.Printf("Connecting to SMTP server: %s (user: %s)", addr, m.config.SMTPUser)
	m.logger.Printf("Full SMTP config before authentication:")
	m.logger.Printf("  Host: %s", m.config.SMTPHost)
	m.logger.Printf("  Port: %s", m.config.SMTPPort)
	m.logger.Printf("  User: %s", m.config.SMTPUser)
	m.logger.Printf("  Password: %s", m.config.SMTPPassword)
	m.logger.Printf("  Password length: %d chars", len(m.config.SMTPPassword))
	m.logger.Printf("  From Address: %s", m.config.FromAddress)

	// Setup authentication
	auth := smtp.PlainAuth("", m.config.SMTPUser, m.config.SMTPPassword, m.config.SMTPHost)
	m.logger.Printf("SMTP authentication configured for host: %s", m.config.SMTPHost)

	// Check if we need to use SSL/TLS (port 465) or STARTTLS (port 587)
	if m.config.SMTPPort == "465" {
		m.logger.Printf("Using SSL/TLS connection for port 465")
		return m.sendEmailSSL(addr, auth, m.config.FromAddress, []string{to}, []byte(message))
	}

	// Send email using standard STARTTLS (port 587 or other)
	m.logger.Printf("Using STARTTLS connection for port %s", m.config.SMTPPort)
	err := smtp.SendMail(addr, auth, m.config.FromAddress, []string{to}, []byte(message))
	if err != nil {
		m.logger.Printf("ERROR: smtp.SendMail failed: %v", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	m.logger.Printf("smtp.SendMail completed successfully for %s", to)
	return nil
}

// sendEmailSSL sends email using SSL/TLS (for port 465)
func (m *Mailer) sendEmailSSL(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Create TLS config
	tlsConfig := &tls.Config{
		ServerName:         m.config.SMTPHost,
		InsecureSkipVerify: false,
	}

	m.logger.Printf("Establishing TLS connection to %s", addr)

	// Connect to the SMTP server with TLS
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		m.logger.Printf("ERROR: TLS dial failed: %v", err)
		return fmt.Errorf("failed to connect via TLS: %w", err)
	}
	defer conn.Close()

	m.logger.Printf("TLS connection established, creating SMTP client")

	// Create SMTP client
	client, err := smtp.NewClient(conn, m.config.SMTPHost)
	if err != nil {
		m.logger.Printf("ERROR: Failed to create SMTP client: %v", err)
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	m.logger.Printf("SMTP client created, authenticating...")
	m.logger.Printf("Authentication details:")
	m.logger.Printf("  Host: %s", m.config.SMTPHost)
	m.logger.Printf("  Port: %s", m.config.SMTPPort)
	m.logger.Printf("  User: %s", m.config.SMTPUser)
	m.logger.Printf("  Password: %s", m.config.SMTPPassword)
	m.logger.Printf("  From: %s", from)

	// Authenticate
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			m.logger.Printf("ERROR: SMTP authentication failed: %v", err)
			m.logger.Printf("Full authentication context:")
			m.logger.Printf("  SMTP Host for auth: %s", m.config.SMTPHost)
			m.logger.Printf("  SMTP User: %s", m.config.SMTPUser)
			m.logger.Printf("  SMTP Password length: %d", len(m.config.SMTPPassword))
			m.logger.Printf("  SMTP Password: %s", m.config.SMTPPassword)
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
		m.logger.Printf("SMTP authentication successful")
	}

	// Set sender
	m.logger.Printf("Setting sender: %s", from)
	if err = client.Mail(from); err != nil {
		m.logger.Printf("ERROR: Failed to set sender: %v", err)
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, addr := range to {
		m.logger.Printf("Adding recipient: %s", addr)
		if err = client.Rcpt(addr); err != nil {
			m.logger.Printf("ERROR: Failed to add recipient %s: %v", addr, err)
			return fmt.Errorf("failed to add recipient: %w", err)
		}
	}

	// Send the email body
	m.logger.Printf("Opening data connection to send message")
	w, err := client.Data()
	if err != nil {
		m.logger.Printf("ERROR: Failed to open data connection: %v", err)
		return fmt.Errorf("failed to open data connection: %w", err)
	}

	m.logger.Printf("Writing message body (%d bytes)", len(msg))
	_, err = w.Write(msg)
	if err != nil {
		m.logger.Printf("ERROR: Failed to write message: %v", err)
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		m.logger.Printf("ERROR: Failed to close data writer: %v", err)
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	m.logger.Printf("Message sent, quitting SMTP session")
	err = client.Quit()
	if err != nil {
		m.logger.Printf("WARNING: Error during QUIT: %v (message may have been sent)", err)
	}

	m.logger.Printf("SUCCESS: Email sent successfully via SSL/TLS to %v", to)
	return nil
}

// LoadConfigFromEnv loads email configuration from environment variables
func LoadConfigFromEnv() Config {
	config := Config{
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromAddress:  getEnv("SMTP_FROM_ADDRESS", "noreply@recontext.online"),
		FromName:     getEnv("SMTP_FROM_NAME", "Recontext"),
	}

	logger := log.New(os.Stdout, "[EMAIL] ", log.LstdFlags)
	logger.Printf("Email configuration loaded:")
	logger.Printf("  SMTP Host: %s", config.SMTPHost)
	logger.Printf("  SMTP Port: %s", config.SMTPPort)
	logger.Printf("  SMTP User: %s", config.SMTPUser)
	logger.Printf("  SMTP Password: %s", maskPassword(config.SMTPPassword))
	logger.Printf("  From Address: %s", config.FromAddress)
	logger.Printf("  From Name: %s", config.FromName)

	return config
}

func maskPassword(password string) string {
	if password == "" {
		return "(empty)"
	}
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + "****" + password[len(password)-2:]
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}
