package main

import (
	"authAPI/users"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	// Регистрируем маршруты
	router.HandleFunc("/auth", users.Auth)           // Бот делает запрос, к нему прикладывает chat_id, затем бот получает ссылку.
	router.HandleFunc("/oauth", users.Oauth_handler) //Это для гитхаба

	http.ListenAndServe(":8080", router)

}