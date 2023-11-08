package users

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type UserData struct {
	Id int64 `json:"id"`
}

var data = make(map[int64]int64)

const (
	CLIENT_ID     = "b783eb30a8893ae6852d"
	CLIENT_SECRET = "660cc24cd1decbfaac1f1fe9793c523ff3954bf1"
)

func Auth(rw http.ResponseWriter, req *http.Request) {
	var tg_id int64
	tg_id, _ = strconv.ParseInt(req.URL.Query().Get("chatid"), 10, 64)
	var str string
	str = strconv.FormatInt(tg_id, 10)
	data[tg_id]++
	log.Print(data[tg_id])
	var authURL string = "https://github.com/login/oauth/authorize?client_id=" + CLIENT_ID + "&state=" + str
	fmt.Fprintf(rw, "%s", authURL)
}

func Oauth_handler(rw http.ResponseWriter, req *http.Request) {
	var responseHtml = "<html><body><h1>Вы не аутентифицированы!</h1></body></html"

	code := req.URL.Query().Get("code") // Достаем временный код из запроса
	tg_id, _ := strconv.ParseInt(req.URL.Query().Get("state"), 10, 64)
	_, ok := data[tg_id]
	log.Print(data[tg_id])
	if code != "" && ok {
		accessToken := getAccessToken(code)
		data[tg_id] = getUserData(accessToken)
		log.Print(data)

		if !checkData(data[tg_id], tg_id) { //Проверяем существует ли док с таким id, если нет, то создаём док.
			register(data[tg_id], tg_id)
		}
		responseHtml = "<html><body><h1>Вы аутентифицированы!</h1></body></html>"

		url := "http://localhost:8081/gitid?githubid=" + strconv.FormatInt(data[tg_id], 10) + "&chatid=" + strconv.FormatInt(tg_id, 10)
		log.Print(url)
		requesturl := fmt.Sprintf(url)
		client := http.Client{}
		request, _ := http.NewRequest("GET", requesturl, nil)
		response, _ := client.Do(request)
		defer response.Body.Close()

		delete(data, tg_id)
	}
	fmt.Fprint(rw, responseHtml)
}

func getAccessToken(code string) string {
	client := http.Client{}
	requestURL := "https://github.com/login/oauth/access_token"
	// Добавляем данные в виде Формы
	form := url.Values{}
	form.Add("client_id", CLIENT_ID)
	form.Add("client_secret", CLIENT_SECRET)
	form.Add("code", code)

	// Готовим и отправляем запрос
	request, _ := http.NewRequest("POST", requestURL, strings.NewReader(form.Encode()))
	request.Header.Set("Accept", "application/json") // просим прислать ответ в формате json
	response, _ := client.Do(request)
	defer response.Body.Close()

	// Достаём данные из тела ответа
	var responsejson struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(response.Body).Decode(&responsejson)
	return responsejson.AccessToken
}

func getUserData(AccessToken string) int64 {
	// Создаём http-клиент с дефолтными настройками
	client := http.Client{}
	requestURL := "https://api.github.com/user"

	// Готовим и отправляем запрос
	request, _ := http.NewRequest("GET", requestURL, nil)
	request.Header.Set("Authorization", "Bearer "+AccessToken)
	response, _ := client.Do(request)
	defer response.Body.Close()

	var data UserData
	json.NewDecoder(response.Body).Decode(&data)
	return data.Id
}
