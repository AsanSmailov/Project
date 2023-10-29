package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	chatids := make([]int64, 0) //масив для хранения chatid,потом исправлю на map когда буду получать id github
	var n int64
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
			find := false
			for _, user := range chatids {
				if user == update.Message.Chat.ID {
					find = true
					break
				}
			}
			if find == false {
				msg.Text = "Привет! Я телеграмм бот c расписанием. \n Чтобы продолжить пользоватьсчя вам нужно зарегестрироваться"
				bot.Send(msg)
				client := http.Client{}
				// Формируем строку запроса вместе с query string
				requestURL := fmt.Sprintf("http://localhost:8080//auth?chatid=pdate.Message.From.UserName.users")
				// Выполняем запрос на сервер. Ответ попадёт в переменную response
				request, _ := http.NewRequest("GET", requestURL, nil)
				response, _ := client.Do(request)
				resBody, _ := io.ReadAll(response.Body) // Получаем тело ответа
				defer response.Body.Close()
				answer := string(resBody)
				msg.Text = answer
				bot.Send(msg)
				//Здесь нужно добавить проверку зарегался ли пользователь
				chatids = append(chatids, n)
				chatids[n] = update.Message.Chat.ID
				n++
				msg.Text = "Вы успешно зарегестрировались"
			} else {
				msg.Text = "Привет! Я телеграмм бот c расписанием. \n Нажми /help чтобы увидеть все команды."
			}

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
		case "c":

		default:
			msg.Text = "Я не понимаю, что вы хотите сказать."
		}

		bot.Send(msg)
	}

}
