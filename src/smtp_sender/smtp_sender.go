package smtp_sender

import (
	"fmt"
	"net/smtp"
)

type SenderParam struct {
	MailBoxUserName string
	MailBoxPass     string
	SmtpHost        string
	SmtpPort        int
	Auth            smtp.Auth
}

func NewMailSender(userName string, pass string, host string, port int) *SenderParam {
	auth := smtp.PlainAuth("", userName, pass, host)
	return &SenderParam{userName, pass, host, port, auth}
}

func (self *SenderParam) SendMsg(to []string, subject string, msg string) error {
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", self.SmtpHost, self.SmtpPort),
		self.Auth,
		self.MailBoxUserName,
		to,
		[]byte(fmt.Sprintf("Subject: %s\n%s", subject, msg)),
	)
	return err
}
