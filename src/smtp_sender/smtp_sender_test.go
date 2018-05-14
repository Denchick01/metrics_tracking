package smtp_sender

import (
	conf "github.com/Denchick01/metrics_tracking/src/configuration"
	"testing"
)

func TestSimpleTest(t *testing.T) {
	var err error

	mainConfig, err := conf.ReadMainConfig("../../etc/config.yaml")
	if err != nil {
		t.Error("Can't read config:", err)
	}

	smtpSender := NewMailSender(
		mainConfig.MailSender.Address,
		mainConfig.MailSender.Password,
		mainConfig.MailSender.SmtpHost,
		mainConfig.MailSender.SmtpPort,
	)

	err = smtpSender.SendMsg(
		[]string{mainConfig.MailSender.Address},
		"Test message",
		"Hello",
	)

	if err != nil {
		t.Error("Can't send message:", err)
	}
}
