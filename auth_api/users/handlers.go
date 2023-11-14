package users

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserData struct {
	Id int64 `json:"id"`
}

// Словарь сессий для авторизации
var data = make(map[int64]int64)

// Данные github oauth
const (
	CLIENT_ID     = "b783eb30a8893ae6852d"
	CLIENT_SECRET = "660cc24cd1decbfaac1f1fe9793c523ff3954bf1"
)

// Handle функция, на запрос с tg id возвращает сгенерированную ссылку на github
func Auth(rw http.ResponseWriter, req *http.Request) {
	//Получаем данные из ссылки
	tg_id, _ := strconv.ParseInt(req.URL.Query().Get("chatid"), 10, 64)
	str := strconv.FormatInt(tg_id, 10)
	//Создаём сессию
	data[tg_id]++
	//Генерируем и отдаём ссылку
	var authURL string = "https://github.com/login/oauth/authorize?client_id=" + CLIENT_ID + "&state=" + str
	fmt.Fprintf(rw, "%s", authURL)
}

// Callback функция для github oauth, возвращает
func Oauth_handler(rw http.ResponseWriter, req *http.Request) {
	var responseHtml = "<html><body><h1>Вы не аутентифицированы!</h1></body></html"

	//Получаем данные из ссылки
	code := req.URL.Query().Get("code") // Достаем временный код из запроса
	tg_id, _ := strconv.ParseInt(req.URL.Query().Get("state"), 10, 64)
	_, ok := data[tg_id]

	//Проверяем наличие временного кода и сессии
	if code != "" && ok {
		//Обмениваем временный код на accesstoken
		accessToken := getAccessToken(code)
		//Получаем github id
		data[tg_id] = getUserData(accessToken)

		if !checkData(data[tg_id], tg_id) { //Проверяем существует ли док с таким id, если нет, то создаём док.
			register(data[tg_id], tg_id)
		}
		responseHtml = "<html><body><h1>Вы аутентифицированы!</h1></body></html>"

		//Создаём запрос callback боту, отправляем ему github id и tg id
		url := "http://localhost:8081/gitid?githubid=" + strconv.FormatInt(data[tg_id], 10) + "&chatid=" + strconv.FormatInt(tg_id, 10)
		requesturl := fmt.Sprintf(url)
		client := http.Client{}
		request, _ := http.NewRequest("GET", requesturl, nil)
		response, _ := client.Do(request)
		defer response.Body.Close()

		delete(data, tg_id)
	}
	fmt.Fprint(rw, responseHtml)
}

// Handle функция, принимает tg id, возвращает bool в зависимости указана ли имя или группа
func CheckAbout(rw http.ResponseWriter, req *http.Request) {
	//Получаем данные из ссылки
	tg_id, _ := strconv.ParseInt(req.URL.Query().Get("chatid"), 10, 64)
	//Получаем структуру user
	user := getData(tg_id, "tg_id")
	// Проверяем наличия имени или группы и отправляем ответ
	if user.About.Group == "" && user.About.FullName == "" {
		fmt.Fprintf(rw, "%t", false)
	} else {
		fmt.Fprintf(rw, "%t", true)
	}
}

// Handle функция, принимает id, информацию, которую надо изменить, тип информации, возвращает bool
func UpdateData(rw http.ResponseWriter, req *http.Request) {
	//Получаем данные
	tg_id, _ := strconv.ParseInt(req.FormValue("chatid"), 10, 64)
	data := req.FormValue("data")
	datatype := req.FormValue("datatype")
	//Отвечаем, используя inputData()
	fmt.Fprintf(rw, "%t", inputData(tg_id, data, datatype))
}

// Handle функция, принимает id, в ответ отдаёт роль пользователя
func GetRole(rw http.ResponseWriter, req *http.Request) {
	//Получаем данные
	tg_id, _ := strconv.ParseInt(req.URL.Query().Get("chatid"), 10, 64)
	//Получаем структуру user
	user := getData(tg_id, "tg_id")
	//Отвечаем
	fmt.Fprintf(rw, "%s", user.Role)
}

// Handle функция, принимает id и действие, возвращает сгенерированный jwt token для расписания
func JWTschedule(rw http.ResponseWriter, req *http.Request) {
	//Получаем данные
	github_id, _ := strconv.ParseInt(req.FormValue("gitid"), 10, 64)
	action := string(req.FormValue("action"))
	//Генерируем секрет
	SECRET := randJWTSecret(16)

	//Отправляем секретный код модулю расписания
	client := http.Client{}
	requesturl := "http://localhost:8082/getSecret"

	form := url.Values{}
	form.Add("SECRET", SECRET)
	request, _ := http.NewRequest("POST", requesturl, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, _ := client.Do(request)
	defer response.Body.Close()

	//Получаем данные
	user := getData(github_id, "github_id")
	line := user.About.Group
	var group, sub_group string
	group = line[:strings.Index(line, "(")]
	sub_group = line[strings.Index(line, "(")+1 : strings.Index(line, ")")]
	log.Print(group, " ", sub_group)
	//Формируем токен
	tokenExpiresAt := time.Now().Add(time.Second * time.Duration(60))
	payload := jwt.MapClaims{
		"action":     action,
		"full_name":  user.About.FullName,
		"group":      group,
		"sub_group":  sub_group,
		"expires_at": tokenExpiresAt,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	//Подписываем токен секретным кодом
	tokenString, err := token.SignedString([]byte(SECRET))
	if err != nil {
		log.Fatal(err)
	}
	log.Print(SECRET)
	fmt.Fprintf(rw, "%s", tokenString)
}

func JWTadmin(rw http.ResponseWriter, req *http.Request) {
	//Получаем данные
	github_id, _ := strconv.ParseInt(req.FormValue("gitid"), 10, 64)
	action := string(req.FormValue("action"))
	//Генерируем секрет
	SECRET := randJWTSecret(16)

	//Отправляем секретный код модулю администрирования
	client := http.Client{}
	requesturl := "http://localhost:8083/getSecret"

	form := url.Values{}
	form.Add("SECRET", SECRET)
	request, _ := http.NewRequest("POST", requesturl, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, _ := client.Do(request)
	defer response.Body.Close()

	//Получаем данные
	user := getData(github_id, "github_id")
	//Формируем токен
	tokenExpiresAt := time.Now().Add(time.Second * time.Duration(60))

	payload := jwt.MapClaims{
		"action":     action,
		"githubID":   user.GithubID,
		"tgID":       user.TgId,
		"full_name":  user.About.FullName,
		"group":      user.About.Group,
		"expires_at": tokenExpiresAt,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	//Подписываем токен секретным кодом
	tokenString, err := token.SignedString([]byte(SECRET))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(rw, "%s", tokenString)
}

// Символы для генерации токена
var charset = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Генерируем случайную строку размером n символов
func randJWTSecret(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func GetAllUsers(rw http.ResponseWriter, req *http.Request) {
	users, err := json.Marshal(giveAllUsers())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(rw, "%s", users)
}

func DeleteUser(rw http.ResponseWriter, req *http.Request) {
	github_id, _ := strconv.ParseInt(req.FormValue("gitid"), 10, 64)
	fmt.Fprintf(rw, "%t", del_user(github_id, "github_id"))
}

// Для Oauth_handler
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
