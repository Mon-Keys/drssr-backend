package main

import (
	"context"
	"drssr/config"
	"drssr/internal/pkg/consts"
	"drssr/internal/pkg/mailer"
	"drssr/internal/pkg/rollback"
	"fmt"
	"log"
	"strconv"
	"strings"

	user_repository "drssr/internal/users/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func main() {
	config.SetConfig()

	logger := logrus.New()

	ctx, rb := rollback.NewCtxRollback(context.Background())

	ur := user_repository.NewPostgresqlRepository(config.Postgres, *logger)

	bot, err := tgbotapi.NewBotAPI(config.TgBotAPIToken.APIToken)
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			actionStr, reqIDStr, found := strings.Cut(update.CallbackQuery.Data, ",")
			if !found {
				logger.Errorf("Invalid data %s", update.CallbackQuery.Data)
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Неизвестный запрос")
				if _, err := bot.Send(msg); err != nil {
					logger.Errorf("Failed to send err msg: %w", err)
					continue
				}
			}

			action, err := strconv.ParseInt(actionStr, 10, 64)
			if err != nil {
				logger.Errorf("Failed to parse action: %w", err)
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Ошибка")
				if _, err := bot.Send(msg); err != nil {
					logger.Errorf("Failed to parse send err msg: %w", err)
					continue
				}
			}

			reqID, err := strconv.ParseUint(reqIDStr, 10, 64)
			if err != nil {
				logger.Errorf("Failed to parse req id: %w", err)
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Ошибка")
				if _, err := bot.Send(msg); err != nil {
					logger.Errorf("Failed to parse send err msg: %w", err)
					continue
				}
			}

			foundReq, err := ur.GetUserStylistRequestByID(ctx, reqID)
			if err != nil {
				logger.Errorf("Failed to found req in db: %w", err)
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Такого запроса не существует")
				if _, err := bot.Send(msg); err != nil {
					logger.Errorf("Failed to parse send err msg: %w", err)
					continue
				}
			}

			user, err := ur.GetUserByID(ctx, foundReq.UID)
			if err != nil {
				logger.Errorf("Failed to found user in db: %w", err)
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Такого пользователя не существует")
				if _, err := bot.Send(msg); err != nil {
					logger.Errorf("Failed to send err msg: %w", err)
					continue
				}
			}

			switch action {
			case consts.TGBotResponseAccept:
				_, err := ur.AcceptStylist(ctx, foundReq.UID)
				if err != nil {
					logger.Errorf("Failed to update user's stylist flag to 'true' in db: %w", err)
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Ошибка")
					if _, err := bot.Send(msg); err != nil {
						logger.Errorf("Failed to send err msg: %w", err)
						continue
					}
				}

				rb.Add(func() {
					_, err := ur.RejectStylist(ctx, foundReq.UID)
					if err != nil {
						logger.Errorf("Failed to rollback accepting of stylist status: %w", err)
					}
				})

				_, err = ur.DeleteStylistRequestByID(ctx, reqID)
				if err != nil {
					rb.Run()

					logger.Errorf("Failed to delete stylist request: %w", err)
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Ошибка")
					if _, err := bot.Send(msg); err != nil {
						logger.Errorf("Failed to send err msg: %w", err)
						continue
					}
				}

				rb.Add(func() {
					_, err := ur.AddStylistRequest(ctx, foundReq.UID)
					if err != nil {
						logger.Errorf("Failed to rollback deleting of old stylist request: %w", err)
					}
				})

				err = mailer.SendAcceptStylistStatus(user)
				if err != nil {
					rb.Run()

					logger.Errorf("Failed to send email for user: %w", err)
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Ошибка")
					if _, err := bot.Send(msg); err != nil {
						logger.Errorf("Failed to send err msg: %w", err)
						continue
					}
				}

				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf(consts.StylistStatusAcceptOKMsg, user.Email))
				if _, err := bot.Send(msg); err != nil {
					logger.Errorf("Failed to send ok msg: %w", err)
					continue
				}
			case consts.TGBotResponseReject:
				_, err = ur.DeleteStylistRequestByID(ctx, reqID)
				if err != nil {
					rb.Run()

					logger.Errorf("Failed to delete stylist request: %w", err)
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Ошибка")
					if _, err := bot.Send(msg); err != nil {
						logger.Errorf("Failed to send err msg: %w", err)
						continue
					}
				}

				rb.Add(func() {
					_, err := ur.AddStylistRequest(ctx, foundReq.UID)
					if err != nil {
						logger.Errorf("Failed to rollback deleting of old stylist request: %w", err)
					}
				})

				err = mailer.SendRejectStylistStatus(user)
				if err != nil {
					rb.Run()

					logger.Errorf("Failed to send email for user: %w", err)
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Ошибка")
					if _, err := bot.Send(msg); err != nil {
						logger.Errorf("Failed to send err msg: %w", err)
						continue
					}
				}

				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf(consts.StylistStatusRejectOKMsg, user.Email))
				if _, err := bot.Send(msg); err != nil {
					logger.Errorf("Failed to send ok msg: %w", err)
					continue
				}
			default:
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Неизвестный запрос")
				if _, err := bot.Send(msg); err != nil {
					logger.Errorf("Failed to send err msg: %w", err)
					continue
				}
			}
		}
	}
}
