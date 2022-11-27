package mailer

import (
	"drssr/config"
	"drssr/internal/models"
	"drssr/internal/pkg/consts"
	"fmt"

	"gopkg.in/gomail.v2"
)

func SendAcceptStylistStatus(user models.User) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", config.Mailer.Email)
	msg.SetHeader("To", user.Email)
	msg.SetHeader("Subject", "Статус стилиста Pose")
	msg.SetBody("text/html", fmt.Sprintf(consts.StylistStatusAcceptMsg, user.Nickname))

	n := gomail.NewDialer("smtp.mail.ru", 465, config.Mailer.Email, config.Mailer.Password)

	if err := n.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send letter: %w", err)
	}
	return nil
}

func SendRejectStylistStatus(user models.User) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", config.Mailer.Email)
	msg.SetHeader("To", user.Email)
	msg.SetHeader("Subject", "Статус стилиста Pose")
	msg.SetBody("text/html", fmt.Sprintf(consts.StylistStatusRejectMsg, user.Nickname))

	n := gomail.NewDialer("smtp.mail.ru", 465, config.Mailer.Email, config.Mailer.Password)

	if err := n.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send letter: %w", err)
	}
	return nil
}
