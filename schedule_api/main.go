package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Day struct { //Структура для добавления дня недели в БД
	Day              string  `bson:"day" json:"day"`
	Week             string  `bson:"week" json:"week"`
	Count_of_lessons int     `bson:"count_of_lessons" json:"count_of_lessons"`
	Subgroup         int     `bson:"subgroup" json:"subgroup"`
	Lesson1          Lessons `bson:"lesson1" json:"lesson1"`
	Lesson2          Lessons `bson:"lesson2" json:"lesson2"`
	Lesson3          Lessons `bson:"lesson3" json:"lesson3"`
	Lesson4          Lessons `bson:"lesson4" json:"lesson4"`
	Lesson5          Lessons `bson:"lesson5" json:"lesson5"`
}
type Lessons struct { //Структура для добавления предмета для дня недели в БД
	Name      string `bson:"name" json:"name"`
	Time      string `bson:"time" json:"time"`
	Type      string `bson:"type" json:"type"`
	Classroom string `bson:"classroom" json:"classroom"`
	Teacher   string `bson:"teacher" json:"teacher"`
	Comment   string `bson:"comment"`
}

var secret string

func handlerSecret(w http.ResponseWriter, r *http.Request) { //Функция получения секретного кода
	secret = r.FormValue("SECRET")
	log.Print(secret)
}

func getSchedule(w http.ResponseWriter, r *http.Request) {
	var action string
	var sub_group string
	var group string
	jwt_string := r.FormValue("jwt")
	token, err := jwt.Parse(jwt_string, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	payload, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		t := int64(payload["expires_at"].(float64))
		log.Print(t)
		if time.Now().Unix() > t {
			log.Fatal(time.Now().Unix())
		} else {
			action = payload["action"].(string)
			sub_group = payload["sub_group"].(string)
			group = payload["group"].(string)
			log.Print(group)
			group = group[strings.Index(group, "-")+1 : len(group)]
			log.Print(group)
		}
	} else {
		log.Fatal(err)
	}
	sub_group_str, _ := strconv.Atoi(sub_group)
	fmt.Fprintf(w, "%s", act(action, group, sub_group_str))
}

func com_to_lesson(w http.ResponseWriter, r *http.Request) {
	var action string
	jwt_string := r.FormValue("jwt")
	num_of_lesson := r.FormValue("num_of_lesson")
	comment := r.FormValue("comment")
	com_subgroup, _ := strconv.Atoi(r.FormValue("subgroup"))
	com_group := r.FormValue("group")
	com_group = com_group[strings.Index(com_group, "-")+1 : len(com_group)]
	token, err := jwt.Parse(jwt_string, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	payload, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		t := int64(payload["expires_at"].(float64))
		log.Print(t)
		if time.Now().Unix() > t {
			log.Fatal(time.Now().Unix())
		} else {
			action = payload["action"].(string)
		}
	} else {
		log.Fatal(err)
	}
	if action == "com_to_lesson" {
		W := week_find()
		LesNum := "lesson" + (num_of_lesson)
		today := conv_day(time.Now().Weekday().String())
		filter := bson.D{{"day", today}, {"subgroup", com_subgroup}, {"week", W}}
		clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
		client, err := mongo.Connect(context.TODO(), clientOptions)
		var result Day
		collection := client.Database("schedule").Collection("PI-" + com_group)
		err = collection.FindOne(context.TODO(), filter).Decode(&result)
		upd := bson.D{{"$set", bson.D{{LesNum + ".comment", comment}}}}
		_, err = collection.UpdateOne(context.TODO(), filter, upd)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, "%s", "Комментарий успешно добавлен!")
	} else {
		fmt.Fprintf(w, "%s", "Ошибка!")
	}
}

func where_group(w http.ResponseWriter, r *http.Request) {
	var action string
	sub_group, _ := strconv.Atoi(r.FormValue("subgroup"))
	group := r.FormValue("group")
	group = group[strings.Index(group, "-")+1 : len(group)]
	jwt_string := r.FormValue("jwt")
	token, err := jwt.Parse(jwt_string, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	payload, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		t := int64(payload["expires_at"].(float64))
		log.Print(t)
		if time.Now().Unix() > t {
			log.Fatal(time.Now().Unix())
		} else {
			action = payload["action"].(string)
		}
	} else {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s", act(action, group, sub_group))
}

func where_teacher(w http.ResponseWriter, r *http.Request) {
	var sub_group string
	var action string
	var group string

	jwt_string := r.FormValue("jwt")
	teacher := r.FormValue("teacher")
	token, err := jwt.Parse(jwt_string, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	payload, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		t := int64(payload["expires_at"].(float64))
		log.Print(t)
		if time.Now().Unix() > t {
			log.Fatal(time.Now().Unix())
		} else {
			action = payload["action"].(string)
			sub_group = payload["sub_group"].(string)
			group = payload["group"].(string)
			group = group[strings.Index(group, "-")+1 : len(group)]
		}
	} else {
		log.Fatal(err)
	}
	if action == "where_teacher" {
		sub, _ := strconv.Atoi(sub_group)
		today := conv_day(time.Now().Weekday().String())
		W := week_find()
		t := time.Now().String()
		t_h, _ := strconv.Atoi(t[11:13])
		t_m, _ := strconv.Atoi(t[14:16])
		result := find_day(today, group, sub, W)
		if (t_h == 8 || (t_h == 9 && t_m < 50)) && (teacher == result.Lesson1.Teacher) {
			fmt.Fprintf(w, "%s", result.Lesson1.Classroom)
		} else if ((t_h == 9 && t_m >= 50) || (t_h == 10) || (t_h == 11 && t_m < 30)) && (teacher == result.Lesson2.Teacher) {
			fmt.Fprintf(w, "%s", result.Lesson2.Classroom)
		} else if ((t_h == 11 && t_m >= 30) || (t_h == 12) || (t_h == 13 && t_m < 20)) && (teacher == result.Lesson3.Teacher) {
			fmt.Fprintf(w, "%s", result.Lesson3.Classroom)
		} else if ((t_h == 13 && t_m >= 20) || (t_h == 14 && t_m < 50)) && (teacher == result.Lesson4.Teacher) {
			fmt.Fprintf(w, "%s", result.Lesson4.Classroom)
		} else if ((t_h == 15) || (t_h == 16 && t_m < 30)) && (teacher == result.Lesson5.Teacher) {
			fmt.Fprintf(w, "%s", result.Lesson5.Classroom)
		} else {
			fmt.Fprintf(w, "%s", "У преподавателя нет пары!")
		}
	}
}

func schedule_update(w http.ResponseWriter, r *http.Request) {
	group := r.FormValue("group")
	Json := r.FormValue("json")

	var J_Day []Day

	filter := bson.D{}
	err := json.Unmarshal([]byte(Json), &J_Day)

	if err != nil {
		log.Fatal(err)
	}
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	collection := client.Database("schedule").Collection("PI-" + group)
	DeleteResult, err := collection.DeleteMany(context.TODO(), filter)
	for i := range J_Day {
		insertResult, err := collection.InsertOne(context.TODO(), J_Day[i])
		if err != nil {
			fmt.Fprintf(w, "%s", err)
			fmt.Fprintf(w, "%s", insertResult)
			fmt.Fprintf(w, "%s", DeleteResult)
		}
	}
}

func conv_day(today string) string {
	switch today {
	case "Monday":
		today = "Понедельник"
	case "Tuesday":
		today = "Вторник"
	case "Wednesday":
		today = "Среда"
	case "Thursday":
		today = "Четверг"
	case "Friday":
		today = "Пятница"
	}
	return today
}

func conv_next_day(today string) string {
	switch today {
	case "Monday":
		today = "Вторник"
	case "Tuesday":
		today = "Среда"
	case "Wednesday":
		today = "Четверг"
	case "Thursday":
		today = "Пятница"
	case "Sunday":
		today = "Понедельник"
	}
	return today
}

func week_find() string {
	var t, w string
	t = time.Now().String()
	m := t[5:7]
	d := t[8:10]
	if ((d == "20" || d == "21" || d == "22" || d == "23" || d == "24" || d == "25" || d == "26") && m == "11") || ((d == "04" || d == "05" || d == "06" || d == "07" || d == "08" || d == "18" || d == "19" || d == "20" || d == "21" || d == "22") && m == "12") {
		w = "нечетная"
	}
	if ((d == "27" || d == "28" || d == "29" || d == "30") && m == "11") || ((d == "01" || d == "11" || d == "12" || d == "13" || d == "14" || d == "15" || d == "25" || d == "26" || d == "27" || d == "28" || d == "29") && m == "12") {
		w = "четная"
	}
	return w
}

func find_day(today string, group string, sub int, W string) Day {
	filter := bson.D{{"day", today}, {"subgroup", sub}, {"week", W}}
	var result Day
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/") //Подключение к БД
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database("schedule").Collection("PI-" + group)
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	return result
}

func resconv(result Day) string {
	var result_str string
	result_str = result.Day + "\n"
	if result.Lesson1.Name != "" {
		result_str += "1 пара: " + result.Lesson1.Name + "\n" + "Тип: " + result.Lesson1.Type + "\n" + "Преподаватель: " + result.Lesson1.Teacher + "\n"
		result_str += "Аудитория: " + result.Lesson1.Classroom + "\n" + "Комментарий преподавателя: " + result.Lesson1.Comment + "\n" + "Время начала: " + result.Lesson1.Time + "\n\n"
	}

	if result.Lesson2.Name != "" {
		result_str += "2 пара: " + result.Lesson2.Name + "\n" + "Тип: " + result.Lesson2.Type + "\n" + "Преподаватель: " + result.Lesson2.Teacher + "\n"
		result_str += "Аудитория: " + result.Lesson2.Classroom + "\n" + "Комментарий преподавателя: " + result.Lesson2.Comment + "\n" + "Время начала: " + result.Lesson2.Time + "\n\n"
	}

	if result.Lesson3.Name != "" {
		result_str += "3 пара: " + result.Lesson3.Name + "\n" + "Тип: " + result.Lesson3.Type + "\n" + "Преподаватель: " + result.Lesson3.Teacher + "\n"
		result_str += "Аудитория: " + result.Lesson3.Classroom + "\n" + "Комментарий преподавателя: " + result.Lesson3.Comment + "\n" + "Время начала: " + result.Lesson3.Time + "\n\n"
	}

	if result.Lesson4.Name != "" {
		result_str += "4 пара: " + result.Lesson4.Name + "\n" + "Тип: " + result.Lesson4.Type + "\n" + "Преподаватель: " + result.Lesson4.Teacher + "\n"
		result_str += "Аудитория: " + result.Lesson4.Classroom + "\n" + "Комментарий преподавателя: " + result.Lesson4.Comment + "\n" + "Время начала: " + result.Lesson4.Time + "\n\n"
	}

	if result.Lesson5.Name != "" {
		result_str += "5 пара: " + result.Lesson5.Name + "\n" + "Тип: " + result.Lesson5.Type + "\n" + "Преподаватель: " + result.Lesson5.Teacher + "\n"
		result_str += "Аудитория: " + result.Lesson5.Classroom + "\n" + "Комментарий преподавателя: " + result.Lesson5.Comment + "\n" + "Время начала: " + result.Lesson5.Time + "\n\n"
	}
	return result_str
}
func act(action string, group string, sub int) string {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/") //Подключение к БД
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	W := week_find()
	switch action {
	case "today_lessons":
		today := conv_day(time.Now().Weekday().String())
		result := find_day(today, group, sub, W)
		return resconv(result)
	case "next_lesson":
		today := conv_day(time.Now().Weekday().String())
		t := time.Now().String()
		t_h, _ := strconv.Atoi(t[11:13])
		t_m, _ := strconv.Atoi(t[14:16])
		result := find_day(today, group, sub, W)
		if t_h < 8 {
			return result.Lesson1.Classroom
		} else if t_h == 8 || (t_h == 9 && t_m < 50) {
			return result.Lesson2.Classroom
		} else if (t_h == 9 && t_m >= 50) || (t_h == 10) || (t_h == 11 && t_m < 30) {
			return result.Lesson3.Classroom
		} else if (t_h == 11 && t_m >= 30) || (t_h == 12) || (t_h == 13 && t_m < 20) {
			return result.Lesson4.Classroom
		} else if (t_h == 13 && t_m >= 20) || (t_h == 14 && t_m < 50) {
			return result.Lesson5.Classroom
		}
	case "tomorrow_lessons":
		today := conv_next_day(time.Now().Weekday().String())
		result := find_day(today, group, sub, W)
		return resconv(result)
	case "monday_lessons":
		today := "Понедельник"
		result := find_day(today, group, sub, W)
		return resconv(result)
	case "tuesday_lessons":
		today := "Вторник"
		result := find_day(today, group, sub, W)
		return resconv(result)
	case "wednesday_lessons":
		today := "Среда"
		result := find_day(today, group, sub, W)
		return resconv(result)
	case "thursday_lessons":
		today := "Четверг"
		result := find_day(today, group, sub, W)
		return resconv(result)
	case "friday_lessons":
		today := "Пятница"
		result := find_day(today, group, sub, W)
		return resconv(result)
	case "where_group":
		today := conv_day(time.Now().Weekday().String())
		t := time.Now().String()
		t_h, _ := strconv.Atoi(t[11:13])
		t_m, _ := strconv.Atoi(t[14:16])
		result := find_day(today, group, sub, W)
		if (t_h == 8 || (t_h == 9 && t_m < 50)) && result.Lesson1.Classroom != "" {
			return result.Lesson1.Classroom
		} else if ((t_h == 9 && t_m >= 50) || (t_h == 10) || (t_h == 11 && t_m < 30)) && result.Lesson2.Classroom != "" {
			return result.Lesson2.Classroom
		} else if ((t_h == 11 && t_m >= 30) || (t_h == 12) || (t_h == 13 && t_m < 20)) && result.Lesson3.Classroom != "" {
			return result.Lesson3.Classroom
		} else if ((t_h == 13 && t_m >= 20) || (t_h == 14 && t_m < 50)) && result.Lesson4.Classroom != "" {
			return result.Lesson4.Classroom
		} else if ((t_h == 15) || (t_h == 16 && t_m < 30)) && result.Lesson5.Classroom != "" {
			return result.Lesson5.Classroom
		}
		return "У подгруппы нет пары!"
	}
	return ""
}

func main() {
	router := mux.NewRouter()                      //Создание роутера
	router.HandleFunc("/getSecret", handlerSecret) //Создание маршрута получения секретного кода
	router.HandleFunc("/getSchedule", getSchedule) //Создание маршрута получения JWT-токена
	router.HandleFunc("/com_to_lesson", com_to_lesson)
	router.HandleFunc("/where_group", where_group)
	router.HandleFunc("/where_teacher", where_teacher)
	router.HandleFunc("/upload_schedule", schedule_update).Methods("POST")
	http.ListenAndServe(":8082", router)
}
