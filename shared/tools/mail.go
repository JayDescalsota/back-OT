package tools

import (
	"log"
	"strconv"

	"gopkg.in/gomail.v2"

	"github.com/clinicmanager/shared/setting"
)

func SendEmail(to, subject, body string) error {

	env, errorEnv := setting.LoadAndValidateEnv([]string{
		"SMTP_FROM",
		"SMTP_PASSWORD",
		"SMTP_HOST",
		"SMTP_PORT",
	})

	if errorEnv != nil {
		log.Fatal("failed to load mail env settings: ", errorEnv)
	}

	from := env["SMTP_FROM"]
	password := env["SMTP_PASSWORD"]
	host := env["SMTP_HOST"]
	portStr := env["SMTP_PORT"]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal("failed to convert port to int: ", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(host, port, from, password)

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
