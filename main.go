package main

import (
	"fmt"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func SetToken(token string) {
	// TODO: здесь типо токен проверяем в бд кто это все дела и добавляем к токенам чела

	fmt.Println(token)
}

func GetHosts(id int) []string {
	// TODO: здесь берем из бд данные

	return []string{"Изотова А.А.", "Бабердин П.В.", "Скляр Л.Н.", "Ковтун И.И.", "Гудкова И.А.", "Гуриков С.Р."}
}

func GetButtonHosts(id int) tgbotapi.InlineKeyboardMarkup {
	var buttons []tgbotapi.KeyboardButton
	hosts := GetHosts(id)

	for _, v := range hosts {
		buttons = append(buttons, tgbotapi.NewKeyboardButton(v))
	}

	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Изотова А.А.", "token1"),
			tgbotapi.NewInlineKeyboardButtonData("Бабердин П.В.", "token2"),
			tgbotapi.NewInlineKeyboardButtonData("Скляр Л.Н.", "token3"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Ковтун И.И.", "token4"),
			tgbotapi.NewInlineKeyboardButtonData("Гудкова И.А.", "token5"),
			tgbotapi.NewInlineKeyboardButtonData("Гуриков С.Р.", "token6"),
		),
	)
	return numericKeyboard
}

func GetHostLocation(token string) {
	// TODO: здесь надо бы получить геопозицю хоста и отослать клиенту

	fmt.Println(token)
}

func TelegramBot() {
	bot, _ := tgbotapi.NewBotAPI("5401428277:AAHAbVaEoKnvPR4IQr0slOu9x2jVLrtTa54")
	bot.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			switch update.Message.Command() {
			case "start":
				msg.ReplyMarkup = GetButtonHosts(update.Message.From.ID)
				msg.Text = "Добро пожаловать!"
				bot.Send(msg)
			case "token":
				token := update.Message.CommandArguments()
				if token == "" {
					msg.Text = "Используйте /token [token]"
					bot.Send(msg)
				} else {
					SetToken(update.Message.CommandArguments())
				}
			default:
				msg.Text = "Неизвестная команда"
				bot.Send(msg)
			}
		} else if update.CallbackQuery != nil {
			GetHostLocation(update.CallbackQuery.Data)
			bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
		}
	}
}

func main() {
	go TelegramBot()

	for {

	}
}
