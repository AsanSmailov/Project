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

// Функция для проверки открытых сессий
func check(chatids map[int64]string, Chat_ID int64) bool {
	for user, _ := range chatids {
		if user == Chat_ID {
			return true
		}
	}
	return false
}

// GET запрос к авторизации для проверки роли
func get_role(Chat_ID int64) string {
	client := http.Client{}
	// Формируем строку запроса вместе с query string
	requestURL := fmt.Sprintf("http://localhost:8080/getRole?chatid=%d", Chat_ID)
	// Выполняем запрос на сервер. Ответ попадёт в переменную response
	request, _ := http.NewRequest("GET", requestURL, nil)
	response, _ := client.Do(request)
	resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
	fmt.Print("ger_role log:", string(resBody))
	defer response.Body.Close()
	return string(resBody)
}

// GET запрос к авторизации для проверки наличия всех данных(ФИО, группа)
func check_data(Chat_ID int64) string {
	client := http.Client{}
	requestURL := fmt.Sprintf("http://localhost:8080/checkAbout?chatid=%d", Chat_ID)
	request, _ := http.NewRequest("GET", requestURL, nil)
	response, _ := client.Do(request)
	resBody, _ := io.ReadAll(response.Body)
	defer response.Body.Close()
	return string(resBody)
}

// GET запрос к авторизации для добовленния данных
func send_data(Chat_ID int64, message string, datatype string) string {
	client := http.Client{}
	requestURL := "http://localhost:8080/updateData"

	form := url.Values{}
	form.Add("chatid", strconv.FormatInt(Chat_ID, 10)) //chat id пользователя которому нужно записать данные
	form.Add("data", message)                          //данные которые нужно записать
	form.Add("datatype", datatype)                     //тип данных (ФИО или группа)

	request, _ := http.NewRequest("POST", requestURL, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, _ := client.Do(request)
	resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
	defer response.Body.Close()
	return string(resBody)
}

// POST запрос к авторизации для ПОЛУЧЕНИЯ JWT TOKEN
func request_jwt(GIT_ID string) string {
	client := http.Client{}
	requestURL := "http://localhost:8080/getJWT/schedule"

	form := url.Values{}
	form.Add("gitid", GIT_ID)

	request, _ := http.NewRequest("POST", requestURL, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, _ := client.Do(request)
	resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
	defer response.Body.Close()
	return string(resBody)
}

func request_jwt_admin(GIT_ID string) string {
	client := http.Client{}
	requestURL := "http://localhost:8080/getJWT/admin"

	form := url.Values{}
	form.Add("github_id", GIT_ID)

	request, _ := http.NewRequest("POST", requestURL, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, _ := client.Do(request)
	resBody, _ := io.ReadAll(response.Body) // Получаем тело ответ
	defer response.Body.Close()
	return string(resBody)
}

func getSchedule(chatids map[int64]string, Chat_ID int64, action string) string {
	client := http.Client{}
	requestURL := fmt.Sprintf("http://localhost:8082//getSchedule")

	form := url.Values{}
	form.Add("jwt", request_jwt(chatids[Chat_ID]))
	form.Add("action", action)

	request, _ := http.NewRequest("POST", requestURL, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, _ := client.Do(request)
	resBody, _ := io.ReadAll(response.Body)
	defer response.Body.Close()
	return string(resBody)
}

func main() {
	//map для хранения открытый сессий
	chatids := make(map[int64]string)
	//создание нового экземпляра бота
	bot, err := tgbotapi.NewBotAPI("6320552220:AAFd90gcVVs0tLR4uXNPzcAL_thmjFAXt4U")
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)
	//Создаем новый объект обновления
	u := tgbotapi.NewUpdate(0)
	//Устанавливаем тайм-аут
	u.Timeout = 60
	//программа настраивает канал обновлений для получения обновлений от бота, используя "bot.GetUpdatesChan(u)"
	updates, _ := bot.GetUpdatesChan(u)
	//Callback для авторизации для получения github_id, если он получин записываем его в map и ваводим сообщение об успешной авторизации
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
		//создание нового сообщения для отправки пользователю в указанный чат
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		//новое сообщение будет ответом на существующее сообщение с ID update.Message.MessageID
		msg.ReplyToMessageID = update.Message.MessageID
		//создание  клавиатуры с кнопками для ответов
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
		if !check(chatids, update.Message.Chat.ID) { //проверяем есть ли для пользоателя открытая сессия
			msg.Text = "Привет! Я телеграмм бот c расписанием. \nЧтобы продолжить пользоваться вам нужно авторизироваться."
			bot.Send(msg)
			//GET запрос для получения ссылки для авторизации
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
			defer response.Body.Close()
		} else {
			if check_data(update.Message.Chat.ID) == "true" {
				switch update.Message.Text {
				case "/start":
					msg.Text = "Привет! Я телеграмм бот c расписанием. \nНажми /help чтобы увидеть все команды."
				case "/help":
					msg.Text = "Список всех команд: \n- Где следующая пара\n- Расписание на день недели\n- Расписание на сегодня\n- Расписание на завтра\n- Оставить комментарий к паре /n- Где группа \n- Где преподаватель\n- toadmin"
				case "toadmin":
					if get_role(update.Message.Chat.ID) == "admin" { //Проверка роли для перехода в админ панель
						client := http.Client{}
						requestURL := fmt.Sprintf("http://localhost:8083//toadmin")
						form := url.Values{}
						form.Add("jwt", request_jwt_admin(chatids[update.Message.Chat.ID]))
						form.Add("gitig", chatids[update.Message.Chat.ID])
						request, _ := http.NewRequest("POST", requestURL, strings.NewReader(form.Encode()))
						request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
						response, _ := client.Do(request)
						resBody, _ := io.ReadAll(response.Body)
						msg.Text = string(resBody)
						defer response.Body.Close()
					} else {
						msg.Text = "Недостаточно прав"
					}
				case "Где следующая пара":
					msg.Text = getSchedule(chatids, update.Message.Chat.ID, "next_lesson")
				case "Расписание на сегодня":
					msg.Text = getSchedule(chatids, update.Message.Chat.ID, "today_lessons")
				case "Расписание на завтра":
					msg.Text = getSchedule(chatids, update.Message.Chat.ID, "tomorrow_lessons")
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
					msg.Text = getSchedule(chatids, update.Message.Chat.ID, "monday_lessons")
				case "Расписание на вторник":
					msg.Text = getSchedule(chatids, update.Message.Chat.ID, "tuesday_lessons")
				case "Расписание на среду":
					msg.Text = getSchedule(chatids, update.Message.Chat.ID, "wednsday_lessons")
				case "Расписание на четверг":
					msg.Text = getSchedule(chatids, update.Message.Chat.ID, "thursday_lessons")
				case "Расписание на пятницу":
					msg.Text = getSchedule(chatids, update.Message.Chat.ID, "friday_lessons")
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
						requestURL := fmt.Sprintf("http://localhost:8082//com_to_lesson")
						form := url.Values{}
						form.Add("jwt", request_jwt(chatids[update.Message.Chat.ID]))
						form.Add("num_of_lesson", num_of_lesson)
						form.Add("group", group)
						request, _ := http.NewRequest("POST", requestURL, strings.NewReader(form.Encode()))
						request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
						response, _ := client.Do(request)
						resBody, _ := io.ReadAll(response.Body)
						msg.Text = string(resBody)
						defer response.Body.Close()
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
						requestURL := fmt.Sprintf("http://localhost:8082//where_group")
						form := url.Values{}
						form.Add("jwt", request_jwt(chatids[update.Message.Chat.ID]))
						form.Add("group", group)
						request, _ := http.NewRequest("POST", requestURL, strings.NewReader(form.Encode()))
						request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
						response, _ := client.Do(request)
						resBody, _ := io.ReadAll(response.Body)
						msg.Text = string(resBody)
						defer response.Body.Close()
					} else {
						msg.Text = "Недостаточно прав"
					}
				case "Где преподаватель":
					teacher := ""
					msg.Text = "Введите ФИО преподавателя"
					bot.Send(msg)
					for update := range updates {
						if update.Message == nil { // ignore any non-Message Updates
							continue
						}
						teacher = update.Message.Text
					}
					client := http.Client{}
					requestURL := fmt.Sprintf("http://localhost:8082//where_teacher")
					form := url.Values{}
					form.Add("jwt", request_jwt(chatids[update.Message.Chat.ID]))
					form.Add("teacher", teacher)
					request, _ := http.NewRequest("POST", requestURL, strings.NewReader(form.Encode()))
					request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					response, _ := client.Do(request)
					resBody, _ := io.ReadAll(response.Body)
					msg.Text = string(resBody)
					defer response.Body.Close()
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
