package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

// declare static directory that holds email templates
// ↓↓↓

//go:embed "email_templates"
var templateFS embed.FS

// mail.Dialer - instance is used to connect to SMPT server
// sender - info (from section)
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	// initialise dialer with SMTP settings
	// 5 second timeout when sending email
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

func (m *Mailer) Send(recipient, templateFile string, data any) error {
	// parse template file in template directory
	tmpl, err := template.New("welcome_email").ParseFS(templateFS, "email_templates/"+templateFile)
	if err != nil {
		return err
	}

	// store {{define "subject"}} into subject bytes buffer
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// store {{define "plainbody"}} into bytes buffer
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainbody", data)
	if err != nil {
		return err
	}

	// same as above
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlbody", data)
	if err != nil {
		return err
	}

	// build email
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// send email
	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}

	return nil
}
