package users

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type UserData struct {
	Id   int64 `json:"id"`
	TgId string
}

var userData UserData

const (
	CLIENT_ID     = "b783eb30a8893ae6852d"
	CLIENT_SECRET = "660cc24cd1decbfaac1f1fe9793c523ff3954bf1"
)

func Auth(rw http.ResponseWriter, req *http.Request) {
	userData.TgId = req.URL.Query().Get("chatid")
	var authURL string = "https://github.com/login/oauth/authorize?client_id=" + CLIENT_ID + "&state=" + userData.TgId
	fmt.Fprintf(rw, "%s", authURL)
}

func Oauth_handler(rw http.ResponseWriter, req *http.Request) {
	var responseHtml = "<html><body><h1>Вы не аутентифицированы!</h1></body></html"

	code := req.URL.Query().Get("code") // Достаем временный код из запроса
	if code != "" {
		accessToken := getAccessToken(code)
		userData.Id = getUserData(accessToken)

		if !checkData(userData.Id, userData.TgId) { //Проверяем существует ли док с таким id, если нет, то создаём док.
			register(userData.Id, userData.TgId)
		}
		responseHtml = "<html><body><h1>Вы аутентифицированы!</h1></body></html>"

		requesturl := fmt.Sprintf("http://localhost:8080/gitid?githubid=%d", userData.Id)
		client := http.Client{}
		request, _ := http.NewRequest("GET", requesturl, nil)
		response, _ := client.Do(request)
		defer response.Body.Close()

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
