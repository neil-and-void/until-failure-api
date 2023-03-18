package mail

import (
	"bytes"
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"
	"text/template"

	"github.com/neilZon/workout-logger-api/config"
)

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unkown fromServer")
		}
	}
	return nil, nil
}

func sendEmail(to []string, subject_line string, body string) error {
	from := string(os.Getenv(config.EMAIL))
	pass := string(os.Getenv(config.APP_PASSWORD))
	auth := LoginAuth(from, pass)

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: " + subject_line + "!\n"
	msg := []byte(subject + mime + "\n" + body)
	addr := "smtp.gmail.com:587"

	if err := smtp.SendMail(addr, auth, from, to, msg); err != nil {
		return err
	}
	return nil
}

func parseTemplate(templateFileName string, data interface{}) (string, error) {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func SendVerificationCode(code string, recipient string) error {
	templateData := struct {
		Link string
	}{
		Link: fmt.Sprintf("http://localhost:8080/verify?code=%s", code),
	}

	abs, err := filepath.Abs("./mail/email-verification-template.html")
	if err != nil {
		return err
	}

	body, err := parseTemplate(abs, templateData)
	if err != nil {
		return err
	}

	err = sendEmail([]string{recipient}, "Email Verification", body)
	if err != nil {
		return err
	}

	return nil
}

func SendResetLink(code string, recipient string) error {
	templateData := struct {
		Link string
	}{
		Link: fmt.Sprintf("tilfailureapp://s?forgotPasswordCode=%s", code),
	}

	abs, err := filepath.Abs("./mail/forgot-password-template.html")
	if err != nil {
		return err
	}

	body, err := parseTemplate(abs, templateData)
	if err != nil {
		return err
	}

	err = sendEmail([]string{recipient}, "Til Failure Password Reset", body)
	if err != nil {
		return err
	}

	return nil
}
