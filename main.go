package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	bot, err := tgbotapi.NewBotAPI("6320552220:AAFd90gcVVs0tLR4uXNPzcAL_thmjFAXt4U")
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ /*err*/ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ReplyToMessageID = update.Message.MessageID

		switch update.Message.Text {
		case "/start":
			msg.Text = "Привет! Я телеграмм бот c расписанием. \n Нажми /help чтобы увидеть все команды."
		case "/help":
			msg.Text = "Список всех команд: ..."
		case "toadmin":
			//if role==admin{
			msg.Text = "ссылка на стр панели администратора"
			//}
		case "Где следующая пара":
			msg.Text = "..."
		case "Расписание на сегодня":
			msg.Text = "..."
		case "Расписание на завтра":
			msg.Text = "..."
		default:
			msg.Text = "Я не понимаю, что вы хотите сказать."
		}

		bot.Send(msg)
	}

}
