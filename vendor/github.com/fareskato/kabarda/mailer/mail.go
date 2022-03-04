package mailer

import (
	"bytes"
	"fmt"
	mail "github.com/ainsleyclark/go-mail"
	"github.com/vanng822/go-premailer/premailer"
	gosimplemail "github.com/xhit/go-simple-mail/v2"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"time"
)

// Mail the mail package type holds all field and uses with all methods
// to send email ....etc
type Mail struct {
	Domain        string
	TemplatesPath string // templates dir path
	Host          string
	Port          int
	UserName      string
	Password      string
	Encryption    string
	FromAddress   string
	FromName      string
	Jobs          chan Message
	Result        chan Result
	Api           string
	ApiKey        string
	ApiUrl        string
}

// Message email message
type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Template    string
	Attachments []string
	Data        interface{}
}

// Result email sending result
type Result struct {
	Success bool
	Error   error
}

func (m *Mail) ListenForEmail() {
	for {
		msg := <-m.Jobs
		err := m.Send(msg)
		if err != nil {
			m.Result <- Result{false, err}
		} else {
			m.Result <- Result{true, nil}
		}
	}
}

// Send will send email vis smtp or other drivers api
func (m *Mail) Send(msg Message) error {
	// API or SMTP ?
	if len(m.Api) > 0 && len(m.ApiKey) > 0 && len(m.ApiUrl) > 0 && m.Api != "smtp" {
		// send via some api like spark, mailgun ...etc
		m.ChooseAndSendViaAPI(msg)
	}
	// or send via smtp
	return m.SendSMTPMessage(msg)
}

// ChooseAndSendViaAPI determine which mail api will be used(mailgun, sparkpost and sendergrid) supported
func (m *Mail) ChooseAndSendViaAPI(msg Message) error {
	switch m.Api {
	case "mailgun", "sparkpost", "sendgrid":
		return m.SendUsingAPI(msg, m.Api)
	default:
		return fmt.Errorf("unknown api: %s; only mailgun or sparkpost or sendergrid accepted", m.Api)
	}
}

// SendUsingAPI send email via mailgun, sparkpost or sendergrid
func (m *Mail) SendUsingAPI(msg Message, transport string) error {
	// ensure that address and name are not empty
	if msg.From == "" {
		msg.From = m.FromAddress
	}
	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	// driver config
	cfg := mail.Config{
		URL:         m.ApiUrl,
		Domain:      m.Domain,
		APIKey:      m.ApiKey,
		FromAddress: m.FromAddress,
		FromName:    m.FromName,
	}

	// get driver
	driver, err := mail.NewClient(transport, cfg)
	if err != nil {
		return err
	}

	// html message
	htmlFormattedMessage, err := m.generateHTMLMessage(msg)
	if err != nil {
		return err
	}

	// plain text message
	plainTextMessage, err := m.generatePlainTextMessage(msg)
	if err != nil {
		return err
	}

	// init Transmission
	tx := &mail.Transmission{
		Recipients: []string{msg.To},
		Subject:    msg.Subject,
		HTML:       htmlFormattedMessage,
		PlainText:  plainTextMessage,
	}

	// add attachments
	err = m.addAPIAttachments(msg, tx)
	if err != nil {
		return err
	}

	// send
	_, err = driver.Send(tx)
	if err != nil {
		return err
	}
	return nil
}

// addAPIAttachments add all attachments to api emails
func (m *Mail) addAPIAttachments(msg Message, tx *mail.Transmission) error {
	if len(msg.Attachments) > 0 {
		var attachments []mail.Attachment
		// loop over all attachments
		for _, x := range msg.Attachments {
			var attach mail.Attachment
			// get the content for each attachment
			content, err := ioutil.ReadFile(x)
			if err != nil {
				return err
			}
			//populate attach
			attachedFileName := filepath.Base(x)
			attach.Filename = attachedFileName
			attach.Bytes = content
			// add all to attachments
			attachments = append(attachments, attach)
		}
		// add attachments to Transmission
		tx.Attachments = attachments
	}
	return nil
}

// SendSMTPMessage send email vis SMTP, this called by Send or can be called directly
// we support formatted html email and plain text email
func (m *Mail) SendSMTPMessage(msg Message) error {
	// ensure that address and name are not empty
	if msg.From == "" {
		msg.From = m.FromAddress
	}
	if msg.FromName == "" {
		msg.FromName = m.FromName
	}
	// html message
	htmlFormattedMessage, err := m.generateHTMLMessage(msg)
	if err != nil {
		return err
	}
	// plain text message
	plainTextMessage, err := m.generatePlainTextMessage(msg)
	if err != nil {
		return err
	}
	// SMTP Server
	server := gosimplemail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.UserName
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	// Variable to keep alive connection
	server.KeepAlive = false
	// Timeout for connect to SMTP Server
	server.ConnectTimeout = 10 * time.Second
	// Timeout for send the data and wait respond
	server.SendTimeout = 10 * time.Second
	// SMTP client
	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}
	// sending email
	email := gosimplemail.NewMSG()
	email.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject)
	email.SetBody(gosimplemail.TextHTML, htmlFormattedMessage)
	email.AddAlternative(gosimplemail.TextPlain, plainTextMessage)
	// email attachments
	if len(msg.Attachments) > 0 {
		for _, atc := range msg.Attachments {
			email.AddAttachment(atc)
		}
	}
	//Pass the client to the email message to send it
	err = email.Send(smtpClient)
	if err != nil {
		return err
	}
	// all good
	return nil
}

// getEncryption convert Mail.Encryption string type to mail.Encryption
func (m *Mail) getEncryption(s string) gosimplemail.Encryption {
	switch s {
	case "tls":
		return gosimplemail.EncryptionSTARTTLS
	case "ssl":
		return gosimplemail.EncryptionSSL
	case "none":
		return gosimplemail.EncryptionNone
	default:
		return gosimplemail.EncryptionSTARTTLS
	}
}

// generateHTMLMessage build html message with css style support
func (m *Mail) generateHTMLMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.html.gohtml", m.TemplatesPath, msg.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}
	formattedMessage := tpl.String()
	formattedMessageWithCss, err := m.inlineCss(formattedMessage)
	if err != nil {
		return "", err
	}
	return formattedMessageWithCss, nil
}

// generatePlainTextMessage build plain text message
func (m *Mail) generatePlainTextMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.plain.gohtml", m.TemplatesPath, msg.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}
	plainMessage := tpl.String()
	return plainMessage, nil
}

// inlineCss add css style support to html format message
func (m *Mail) inlineCss(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}
	p, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}
	html, err := p.Transform()
	if err != nil {
		return "", err
	}
	return html, nil
}
