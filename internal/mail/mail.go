package mail

import (
	"bytes"
	"embed"
	"html/template"
	"os"

	"github.com/wneessen/go-mail"
)

//go:embed "templates"
var templateFS embed.FS

func PrepareEmail(recipient, templateFile string, data interface{}) (*mail.Msg, error) {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return nil, err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return nil, err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return nil, err
	}

	m := mail.NewMsg()
	m.From(os.Getenv("SMTP_SENDER"))
	m.To(recipient)
	m.Subject(subject.String())
	m.SetBodyString(mail.TypeTextHTML, htmlBody.String())

	return m, nil
}
