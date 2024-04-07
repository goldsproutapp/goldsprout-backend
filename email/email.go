package email

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"strconv"

	"github.com/patrickjonesuk/investment-tracker/config"
	"github.com/patrickjonesuk/investment-tracker/models"
	"github.com/patrickjonesuk/investment-tracker/util"
	"github.com/wneessen/go-mail"
)

func Client() *mail.Client {
	client, _ := mail.NewClient(
		config.RequiredEnv(ENVKEY_SMTP_HOST),
		mail.WithPort(
			util.ParseIntOrDefault(config.EnvOrDefault(ENVKEY_SMTP_PORT,
				strconv.FormatInt(DEFAULT_SMTP_PORT, 10)), DEFAULT_SMTP_PORT)),
		mail.WithSMTPAuth(mail.SMTPAuthLogin),
		mail.WithUsername(config.RequiredEnv(ENVKEY_SMTP_USER)),
		mail.WithPassword(config.RequiredEnv(ENVKEY_SMTP_PASS)),
	)
	return client
}

func newMessage(to string, subject string) *mail.Msg {
	msg := mail.NewMsg()
	msg.To(to)
	msg.From(config.EnvOrDefault(ENVKEY_SMTP_FROM, config.RequiredEnv(ENVKEY_SMTP_USER)))
	msg.Subject(subject)
	return msg
}

func SendPlainText(to string, subject string, content string) {
	msg := newMessage(to, subject)
	msg.SetBodyString(mail.TypeTextPlain, content)

	err := SendMessage(msg)
	if err != nil {
        fmt.Println("Error sending email: " + err.Error())
		// TODO: proper log message
	}
}

func SendMessage(message *mail.Msg) error {
	client := Client()
	err := client.DialAndSend(message)
	return err
}

func TemplateFile(name string) *template.Template {
	tmpl, _ := template.ParseFiles("templates/" + name + ".html")
	return tmpl
}

func SendInvitation(to string, by models.User, token string) bool {
	msg := newMessage(to, "Invitation to track your investments")
	url := fmt.Sprintf("%s/invitation?t=%s&e=%s",
		config.RequiredEnv(config.FRONTEND_BASE_URL),
		token,
		base64.StdEncoding.EncodeToString([]byte(to)),
	)
	msg.SetBodyHTMLTemplate(TemplateFile("invitation"), map[string]string{
		"Name":      by.Name(),
		"Email":     by.Email,
		"AcceptURL": url,
	})
	err := SendMessage(msg)
	return err == nil
}
