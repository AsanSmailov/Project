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
	router.HandleFunc("/updateData", users.UpdateData).Methods("POST")       //Изменение данных пользователя. Функция принимает tg_id, данные, тип данных(role, group, full_name)
	router.HandleFunc("/getRole", users.GetRole)                             //Отправка роли пользователя по его tg_id
	router.HandleFunc("/getJWT/schedule", users.JWTschedule).Methods("POST") //Формирование jwt token для модуля Расписания
	router.HandleFunc("/getJWT/admin", users.JWTadmin).Methods("POST")       //Формирование jwt token для модуля Администрирования

	router.HandleFunc("/getAllUsers", users.GetAllUsers).Methods("POST") //В json формате передаёт информацию о всех пользователей
	router.HandleFunc("/DeleteUser", users.DeleteUser).Methods("POST")   //Удаление пользователя по его gihub_id
	http.ListenAndServe(":8080", router)

}
