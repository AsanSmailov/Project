package main

import (
	"authAPI/users"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	// Регистрируем маршруты
	router.HandleFunc("/auth", users.Auth)                                   //Формирование ссылки github
	router.HandleFunc("/oauth", users.Oauth_handler)                         //github callback
	router.HandleFunc("/checkAbout", users.CheckAbout)                       //Проверяет наличия имени или группы
	router.HandleFunc("/sendAbout", users.SendAbout).Methods("POST")         //Изменение имени или группы
	router.HandleFunc("/getRole", users.GetRole)                             //Отправка роли пользователя
	router.HandleFunc("/getJWT/schedule", users.JWTschedule).Methods("POST") //Формирование jwt token для модуля Расписания
	router.HandleFunc("/getJWT/admin", users.JWTadmin).Methods("POST")
	http.ListenAndServe(":8080", router)

}
