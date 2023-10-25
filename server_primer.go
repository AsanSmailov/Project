package main

import (
	"fmt"
	"net/http"
	"strconv"
)

// для сервера
// для преобразования строка->число

func main() {
	http.HandleFunc("/sum", handler)  // Обработчик отвечающий на запроса к /sum
	http.ListenAndServe(":8080", nil) // Запуск сервера на порту 8080
}

func sum(a, b int) int {
	return a + b
}

// Функция которая будет вызвана обработчиком, когда придёт запрос
func handler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем данные из query string и преобразуем в целые числа
	a, _ := strconv.Atoi(r.URL.Query().Get("a"))
	b, _ := strconv.Atoi(r.URL.Query().Get("b"))

	// Собственно сумма
	c := sum(a, b)

	// Записываем текст в ответ
	fmt.Fprintf(w, "%d", c)
}
