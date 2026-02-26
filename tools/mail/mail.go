package mail

import (
	"net/smtp"
	"strconv"
	"strings"

	"github.com/unluckythoughts/go-microservice/v2/utils"
)

type Service struct {
	from string
	addr string
	auth smtp.Auth
}

type Options struct {
	// Host is the SMTP server host. It defaults to "smtp.gmail.com" if not provided.
	Host string `env:"MAIL_HOST" envDefault:"smtp.gmail.com"`
	// Port is the SMTP server port. It defaults to 587 if not provided.
	Port int `env:"MAIL_PORT" envDefault:"587"`
	// Encryption specifies the encryption method to use when connecting to the SMTP server.
	// It can be "tls", "ssl", or "none". It defaults to "tls" if not provided.
	Encryption string `env:"MAIL_ENCRYPTION" envDefault:"tls"`

	// Username is the email address that will be used to authenticate with the SMTP server.
	// It should be set in the environment variable MAIL_USERNAME.
	//
	// This email address will also be used as the "From" address when sending emails, so it should be a valid email address that you have access to.
	Username string `env:"MAIL_USERNAME" envDefault:""`
	// AppPassword is the application-specific password for the email account, required for authentication with the SMTP server.
	// It should be set in the environment variable MAIL_APP_PASSWORD.
	AppPassword string `env:"MAIL_APP_PASSWORD" envDefault:""`
}

type Email struct {
	// To is a slice of recipient email addresses. It must contain at least one valid email address.
	To []string `valid:"emails~at least one valid recipient email is required"`
	// Subject is the subject of the email. It is required and cannot be empty.
	Subject string `valid:"required~email subject is required"`
	// Body is the content of the email. It is required and cannot be empty.
	// The body can contain HTML content, and the mail service will set the appropriate MIME type for HTML emails.
	Body string `valid:"required~email body is required"`
}

func defaultOptions(overrides *Options) *Options {
	opts := Options{
		Host:       "smtp.gmail.com",
		Port:       587,
		Encryption: "tls",
	}

	utils.ParseEnvironmentVars(&opts)

	if overrides.Host != "" {
		opts.Host = overrides.Host
	}
	if overrides.Port != 0 {
		opts.Port = overrides.Port
	}
	if overrides.Encryption != "" {
		opts.Encryption = overrides.Encryption
	}
	if overrides.Username != "" {
		opts.Username = overrides.Username
	}
	if overrides.AppPassword != "" {
		opts.AppPassword = overrides.AppPassword
	}

	return &opts
}

func New(opts *Options) (*Service, error) {
	opts = defaultOptions(opts)

	addr := opts.Host + ":" + strconv.Itoa(opts.Port)

	var auth smtp.Auth
	username := strings.TrimSpace(opts.Username)
	if username != "" {
		auth = smtp.PlainAuth("", username, opts.AppPassword, opts.Host)
	}

	return &Service{
		from: opts.Username,
		addr: addr,
		auth: auth,
	}, nil
}

func (s *Service) SendEmail(e *Email) error {
	if err := utils.ValidateStruct(e); err != nil {
		return err
	}

	body := e.Body
	if !strings.Contains(body, "<body>") {
		body = strings.ReplaceAll(body, "\r\n", "\n")
		body = strings.ReplaceAll(body, "\n", "<br>")
		body = "<html><body>" + body + "</body></html>"
	}

	msg := []byte(strings.Join([]string{
		"From: " + s.from,
		"To: " + strings.Join(e.To, ", "),
		"Subject: " + e.Subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"UTF-8\"",
		"",
		body,
	}, "\r\n"))

	return smtp.SendMail(s.addr, s.auth, s.from, e.To, msg)
}
