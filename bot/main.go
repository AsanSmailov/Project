package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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

func get_role(Chat_ID int64) string {
	client := http.Client{}
	// Формируем строку запроса вместе с query string
	requestURL := fmt.Sprintf("http://localhost:8080/getRole?chatid=%d", Chat_ID)
	// Выполняем запрос на сервер. Ответ попадёт в переменную response
	request, _ := http.NewRequest("GET", requestURL, nil)
	response, _ := client.Do(request)
	resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
	fmt.Print("ger_role log:", string(resBody))
	return string(resBody)
}
func check_data(Chat_ID int64) string {
	client := http.Client{}
	requestURL := fmt.Sprintf("http://localhost:8080/checkAbout?chatid=%d", Chat_ID)
	request, _ := http.NewRequest("GET", requestURL, nil)
	response, _ := client.Do(request)
	resBody, _ := io.ReadAll(response.Body)
	return string(resBody)
}

func send_data(Chat_ID int64, message string, datatype string) string {
	client := http.Client{}
	requestURL := "http://localhost:8080/sendAbout"

	form := url.Values{}
	form.Add("chatid", strconv.FormatInt(Chat_ID, 10))
	form.Add("data", message)
	form.Add("datatype", datatype)

	request, _ := http.NewRequest("POST", requestURL, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, _ := client.Do(request)
	resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ

	return string(resBody)
}

func main() {
	chatids := make(map[int64]string)
	bot, err := tgbotapi.NewBotAPI("6320552220:AAFd90gcVVs0tLR4uXNPzcAL_thmjFAXt4U")
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

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
	go func() { http.ListenAndServe(":8081", nil) }() // Запуск сервера на порту 8081

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ReplyToMessageID = update.Message.MessageID
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Где следующая пара"),
				tgbotapi.NewKeyboardButton("Расписание на сегодня"),
				tgbotapi.NewKeyboardButton("Расписание на завтра"),
				tgbotapi.NewKeyboardButton("Расписание на дни недели"),
				tgbotapi.NewKeyboardButton("Где преподаватель"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Изменить данные(ФИО, группа)"),
				tgbotapi.NewKeyboardButton("Оставить комментарий к паре"),
				tgbotapi.NewKeyboardButton("Где группа"),
				tgbotapi.NewKeyboardButton("toadmin"),
				tgbotapi.NewKeyboardButton("Выйти"),
			),
		)
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
		} else {
			if check_data(update.Message.Chat.ID) == "true" {
				switch update.Message.Text {
				case "/start":
					msg.Text = "Привет! Я телеграмм бот c расписанием. \nНажми /help чтобы увидеть все команды."
				case "/help":
					msg.Text = "Список всех команд: ..."
				case "toadmin":
					if get_role(update.Message.Chat.ID) == "admin" { //Проверка роли для перехода в админ панель
						msg.Text = "ссылка на стр панели администратора"
					} else {
						msg.Text = "Недостаточно прав"
					}
				case "Где следующая пара":
					client := http.Client{}
					requestURL := fmt.Sprintf("http://localhost:8082//next_lesson?chatid=%d", update.Message.Chat.ID)
					request, _ := http.NewRequest("GET", requestURL, nil)
					response, _ := client.Do(request)
					resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
					msg.Text = string(resBody)
				case "Расписание на сегодня":
					client := http.Client{}
					requestURL := fmt.Sprintf("http://localhost:8082//today_lessons?chatid=%d", update.Message.Chat.ID)
					request, _ := http.NewRequest("GET", requestURL, nil)
					response, _ := client.Do(request)
					resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
					msg.Text = string(resBody)
				case "Расписание на завтра":
					client := http.Client{}
					requestURL := fmt.Sprintf("http://localhost:8082//tomorrow_lessons?chatid=%d", update.Message.Chat.ID)
					request, _ := http.NewRequest("GET", requestURL, nil)
					response, _ := client.Do(request)
					resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
					msg.Text = string(resBody)
				case "Расписание на дни недели":
					newKeyboard := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите день недели:")
					newKeyboard.ReplyMarkup = tgbotapi.NewReplyKeyboard(
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Расписание на понедельник"),
							tgbotapi.NewKeyboardButton("Расписание на вторник"),
							tgbotapi.NewKeyboardButton("Расписание на среду"),
							tgbotapi.NewKeyboardButton("Расписание на четверг"),
						),
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Расписание на пятницу"),
							tgbotapi.NewKeyboardButton("Расписание на субботу"),
							tgbotapi.NewKeyboardButton("Расписание на воскресенье"),
							tgbotapi.NewKeyboardButton("Назад"),
						),
					)
					bot.Send(newKeyboard)
				case "Расписание на понедельник":
					client := http.Client{}
					requestURL := fmt.Sprintf("http://localhost:8082//monday_lessons?chatid=%d", update.Message.Chat.ID)
					request, _ := http.NewRequest("GET", requestURL, nil)
					response, _ := client.Do(request)
					resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
					msg.Text = string(resBody)
				case "Расписание на вторник":
					client := http.Client{}
					requestURL := fmt.Sprintf("http://localhost:8082//tuesday_lessons?chatid=%d", update.Message.Chat.ID)
					request, _ := http.NewRequest("GET", requestURL, nil)
					response, _ := client.Do(request)
					resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
					msg.Text = string(resBody)
				case "Расписание на среду":
					client := http.Client{}
					requestURL := fmt.Sprintf("http://localhost:8082//wednsday_lessons?chatid=%d", update.Message.Chat.ID)
					request, _ := http.NewRequest("GET", requestURL, nil)
					response, _ := client.Do(request)
					resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
					msg.Text = string(resBody)
				case "Расписание на четверг":
					client := http.Client{}
					requestURL := fmt.Sprintf("http://localhost:8082//thursday_lessons?chatid=%d", update.Message.Chat.ID)
					request, _ := http.NewRequest("GET", requestURL, nil)
					response, _ := client.Do(request)
					resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
					msg.Text = string(resBody)
				case "Расписание на пятницу":
					client := http.Client{}
					requestURL := fmt.Sprintf("http://localhost:8082//friday_lessons?chatid=%d", update.Message.Chat.ID)
					request, _ := http.NewRequest("GET", requestURL, nil)
					response, _ := client.Do(request)
					resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
					msg.Text = string(resBody)
				case "Расписание на субботу":
					msg.Text = "Выходной. В данный день пар нет."
				case "Расписание на воскресенье":
					msg.Text = "Выходной. В данный день пар нет."
				case "Выйти":
					delete(chatids, update.Message.Chat.ID)
					msg.Text = "Вы успешно вышли!"
				case "Оставить комментарий к паре":
					if get_role(update.Message.Chat.ID) == "teacher" { //проверка роли для добавления коментария к паре
						num_of_lesson := ""
						msg.Text = "Введите номер пары"
						bot.Send(msg)
						for update := range updates {
							if update.Message == nil { // ignore any non-Message Updates
								continue
							}
							num_of_lesson = update.Message.Text
						}
						group := ""
						msg.Text = "Введите номер группы"
						bot.Send(msg)
						for update := range updates {
							if update.Message == nil { // ignore any non-Message Updates
								continue
							}
							group = update.Message.Text
						}
						client := http.Client{}
						requestURL := fmt.Sprintf("http://localhost:8082//com_to_lesson?num_of_lesson=%s&group=%s", num_of_lesson, group)
						request, _ := http.NewRequest("GET", requestURL, nil)
						response, _ := client.Do(request)
						resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
						msg.Text = string(resBody)
					} else {
						msg.Text = "Недостаточно прав"
					}
				case "Где группа":
					if get_role(update.Message.Chat.ID) == "teacher" { //проверка роли
						group := ""
						msg.Text = "Введите номер группы"
						bot.Send(msg)
						for update := range updates {
							if update.Message == nil { // ignore any non-Message Updates
								continue
							}
							group = update.Message.Text
						}
						client := http.Client{}
						requestURL := fmt.Sprintf("http://localhost:8082//where_group?group=%s", group)
						request, _ := http.NewRequest("GET", requestURL, nil)
						response, _ := client.Do(request)
						resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
						msg.Text = string(resBody)
					} else {
						msg.Text = "Недостаточно прав"
					}
				case "Где преподаватель":
					teacher := ""
					msg.Text = "Введите номер группы"
					bot.Send(msg)
					for update := range updates {
						if update.Message == nil { // ignore any non-Message Updates
							continue
						}
						teacher = update.Message.Text
					}
					client := http.Client{}
					requestURL := fmt.Sprintf("http://localhost:8082//where_teacher?teacher=%s", teacher)
					request, _ := http.NewRequest("GET", requestURL, nil)
					response, _ := client.Do(request)
					resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
					msg.Text = string(resBody)
				case "Изменить данные(ФИО, группа)":
					msg.Text = "Отправте ваше ФИО (прим. Иванов Иван Иванович)"
					bot.Send(msg)
					for update := range updates {
						if update.Message == nil { // ignore any non-Message Updates
							continue
						}
						if send_data(update.Message.Chat.ID, update.Message.Text, "full_name") == "true" {
							msg.Text = "Данные успешно записаны."
							bot.Send(msg)
							msg.Text = ""
							break
						} else {
							msg.Text = "Ошибка! Не удалось записать данные."
							bot.Send(msg)
							msg.Text = ""
						}
					}
					msg.Text = "Отправте вашу группу (прим. ИВТ-123(1), ПМИ-123(2), ПИ-123(1) и т.д)"
					bot.Send(msg)
					for update := range updates {
						if send_data(update.Message.Chat.ID, update.Message.Text, "group") == "true" {
							msg.Text = "Данные успешно записаны."
							bot.Send(msg)
							msg.Text = ""
							break
						} else {
							msg.Text = "Ошибка! Не удалось записать данные."
							bot.Send(msg)
							msg.Text = ""
						}
					}
				case "Назад":
					msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Где следующая пара"),
							tgbotapi.NewKeyboardButton("Расписание на сегодня"),
							tgbotapi.NewKeyboardButton("Расписание на завтра"),
							tgbotapi.NewKeyboardButton("Расписание на дни недели"),
							tgbotapi.NewKeyboardButton("Где преподаватель"),
						),
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton("Изменить данные(ФИО, группа)"),
							tgbotapi.NewKeyboardButton("Оставить комментарий к паре"),
							tgbotapi.NewKeyboardButton("Где группа"),
							tgbotapi.NewKeyboardButton("toadmin"),
							tgbotapi.NewKeyboardButton("Выйти"),
						),
					)
				default:
					msg.Text = "Я не понимаю, что вы хотите сказать."
				}
			} else {
				msg.Text = "Необходимо отправить данные!"
				bot.Send(msg)
				msg.Text = "Отправте ваше ФИО (прим. Иванов Иван Иванович)"
				bot.Send(msg)
				for update := range updates {
					if update.Message == nil { // ignore any non-Message Updates
						continue
					}
					if send_data(update.Message.Chat.ID, update.Message.Text, "full_name") == "true" {
						msg.Text = "Данные успешно записаны."
						bot.Send(msg)
						msg.Text = ""
						break
					} else {
						msg.Text = "Ошибка! Не удалось записать данные."
						bot.Send(msg)
						msg.Text = ""
					}
				}
				msg.Text = "Отправте вашу группу (прим. ИВТ-123(1), ПМИ-123(2), ПИ-123(1) и т.д)"
				bot.Send(msg)
				for update := range updates {
					if send_data(update.Message.Chat.ID, update.Message.Text, "group") == "true" {
						msg.Text = "Данные успешно записаны."
						bot.Send(msg)
						msg.Text = ""
						break
					} else {
						msg.Text = "Ошибка! Не удалось записать данные."
						bot.Send(msg)
						msg.Text = ""
					}
				}
			}
		}
		bot.Send(msg)
	}
}
