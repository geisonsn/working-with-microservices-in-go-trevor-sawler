package main

import (
	"bytes"
	"log"
	"text/template"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	Html        *Html
	PlainText   *PlainText
}

type MailBuilder interface {
	build(msg Message) (string, error)
	format(msg string) error
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
}

func (m *Mail) SendSMTPMessage(msg Message) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	data := map[string]any{
		"message": msg.Data,
	}

	msg.DataMap = data

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject)

	err = m.buildMessage(msg)
	if err != nil {
		log.Println(err)
		return err
	}

	email.SetBody(mail.TextPlain, m.PlainText.message)
	email.AddAlternative(mail.TextHTML, m.Html.message)

	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (m *Mail) buildMessage(msg Message) error {
	m.PlainText = &PlainText{template: "./templates/mail.plain.gohtml"}
	m.Html = &Html{template: "./templates/mail.html.gohtml"}

	if err := m.PlainText.build(msg); err != nil {
		log.Println(err)
		return err
	}

	if err := m.Html.build(msg); err != nil {
		log.Println(err)
		return err
	}

	return nil

}

type Html struct {
	template string
	message  string
}

func (m *Html) build(msg Message) error {
	t, err := template.New("email-html").ParseFiles(m.template)
	if err != nil {
		return err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return err
	}

	message := tpl.String()
	err = m.format(message)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil

}

func (m *Html) format(msg string) error {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(msg, &options)
	if err != nil {
		log.Println(err)
		return err
	}

	html, err := prem.Transform()
	if err != nil {
		log.Println(err)
		return err
	}

	m.message = html

	return nil
}

type PlainText struct {
	template string
	message  string
}

func (m *PlainText) build(msg Message) error {
	t, err := template.New("email-plain").ParseFiles(m.template)
	if err != nil {
		return err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		log.Println(err)
		return err
	}

	if err = m.format(tpl.String()); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (m *PlainText) format(msg string) error {

	m.message = msg

	return nil

}

func (m *Mail) getEncryption(s string) mail.Encryption {
	switch s {
	case "tls":
		return mail.EncryptionSTARTTLS

	case "ssl":
		return mail.EncryptionSSLTLS
	case "none", "":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
