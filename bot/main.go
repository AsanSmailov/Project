package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func check(chatids map[int64]string, Chat_ID int64) bool {
	for user, _ := range chatids {
		if user == Chat_ID {
			return true
		}
	}
	return false
}

func main() {
	chatids := make(map[int64]string)
	//chatids := make([]int64, 0) //масив для хранения chatid,потом исправлю на map когда буду получать id github
	bot, err := tgbotapi.NewBotAPI("6320552220:AAFd90gcVVs0tLR4uXNPzcAL_thmjFAXt4U")
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ /*err*/ := bot.GetUpdatesChan(u)

	http.HandleFunc("/gitid", func(w http.ResponseWriter, r *http.Request) { // Обработчик отвечающий на запроса к /gitid
		log.Printf("github_id:.")
		github_id := r.URL.Query().Get("githubid")
		chat_id, _ := strconv.ParseInt(r.URL.Query().Get("chatid"), 10, 64)
		log.Printf("github_id: %s", github_id)
		if github_id != "" {
			chatids[chat_id] = github_id
			bot.Send(tgbotapi.NewMessage(chat_id, "Вы успешно авторизировались!"))
		}
	})
	go func() { http.ListenAndServe(":8081", nil) }() // Запуск сервера на порту 8080

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ReplyToMessageID = update.Message.MessageID

		if !check(chatids, update.Message.Chat.ID) {
			msg.Text = "Привет! Я телеграмм бот c расписанием. \nЧтобы продолжить пользоваться вам нужно авторизироваться."
			bot.Send(msg)
			client := http.Client{}
			// Формируем строку запроса вместе с query string
			requestURL := fmt.Sprintf("http://localhost:8080//auth?chatid=%d", update.Message.Chat.ID)
			// Выполняем запрос на сервер. Ответ попадёт в переменную response
			request, _ := http.NewRequest("GET", requestURL, nil)
			response, _ := client.Do(request)
			resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
			msg.Text = string(resBody)
			bot.Send(msg)
			msg.Text = ""
			//defer response.Body.Close()
			//Здесь нужно добавить проверку зарегался ли пользователь
			/* requestURL = "http://localhost:8080//gitid?"
			request, _ = http.NewRequest("GET", requestURL, nil)
			response, _ = client.Do(request)
			resBody = nil
			resBody, err = io.ReadAll(response.Body)
			log.Printf(string(resBody))
			if err == nil {
				chatids[update.Message.Chat.ID] = string(resBody)
				msg.Text = "Вы успешно авторизировались!"
			} */
		} else {
			switch update.Message.Text {
			case "/start":
				msg.Text = "Привет! Я телеграмм бот c расписанием. \nНажми /help чтобы увидеть все команды."
			case "/help":
				if check(chatids, update.Message.Chat.ID) {
					msg.Text = "Список всех команд: ..."
				}
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
		}

		bot.Send(msg)
	}

}
