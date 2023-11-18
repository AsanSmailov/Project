package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	//"net/http"
	"time"

	//"github.com/golang-jwt/jwt/v5"
	//"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Day struct { //Структура для добавления дня недели в БД
	Day              string
	Week             string
	Count_of_lessons int
	Subgroup         int
	Lesson1          Lessons
	Lesson2          Lessons
	Lesson3          Lessons
	Lesson4          Lessons
	Lesson5          Lessons
}
type Lessons struct { //Структура для добавления предмета для дня недели в БД
	Name      string
	Time      string
	Type      string
	Classroom string
	Teacher   string
	Comment   string
}

var secret string

/*func handlerSecret(w http.ResponseWriter, r *http.Request) { //Функция получения секретного кода
	secret = r.FormValue("SECRET")
	log.Print(secret)
}

func getSchedule(w http.ResponseWriter, r *http.Request) {
	var action string
	var full_name string
	var group string
	var sub_group string
	jwt_string := r.FormValue("jwt")
	token, err := jwt.Parse(jwt_string, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	payload, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		t := int64(payload["expires_at"].(float64))
		if time.Now().Unix() > t {
			log.Fatal(err)
		} else {
			action = payload["action"].(string)
			full_name = payload["full_name"].(string)
			group = payload["group"].(string)
			sub_group = payload["sub_group"].(string)
		}
	} else {
		log.Fatal(err)
	}
	log.Print(action)
	log.Print(full_name)
	log.Print(group)
	log.Print(sub_group)
}*/

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
	if ((d == "20" || d == "21" || d == "22" || d == "23" || d == "24") && m == "11") || ((d == "04" || d == "05" || d == "06" || d == "07" || d == "08" || d == "18" || d == "19" || d == "20" || d == "21" || d == "22") && m == "12") {
		w = "нечетная"
	}
	if ((d == "27" || d == "28" || d == "29" || d == "30") && m == "11") || ((d == "01" || d == "11" || d == "12" || d == "13" || d == "14" || d == "15" || d == "25" || d == "26" || d == "27" || d == "28" || d == "29") && m == "12") {
		w = "четная"
	}
	return w
}

func find_day(today string, sub int, W string) Day {
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
	collection := client.Database("schedule").Collection("PI-232")
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	return result
}

func act(action string, sub int) {
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
	case "lessons_today":
		today := conv_day(time.Now().Weekday().String())
		result := find_day(today, sub, W)
		fmt.Print(result)
	case "next_lesson":
		today := conv_day(time.Now().Weekday().String())
		t := time.Now().String()
		t_h, _ := strconv.Atoi(t[11:13])
		t_m, _ := strconv.Atoi(t[14:16])
		result := find_day(today, sub, W)
		if t_h < 8 {
			fmt.Print(result.Lesson1.Classroom)
		} else if t_h == 8 || (t_h == 9 && t_m < 50) {
			fmt.Print(result.Lesson2.Classroom)
		} else if (t_h == 9 && t_m >= 50) || (t_h == 10) || (t_h == 11 && t_m < 30) {
			fmt.Print(result.Lesson3.Classroom)
		} else if (t_h == 11 && t_m >= 30) || (t_h == 12) || (t_h == 13 && t_m < 20) {
			fmt.Print(result.Lesson4.Classroom)
		} else if (t_h == 13 && t_m >= 20) || (t_h == 14 && t_m < 50) {
			fmt.Print(result.Lesson5.Classroom)
		}
	case "tomorrow_lessons":
		today := conv_next_day(time.Now().Weekday().String())
		result := find_day(today, sub, W)
		fmt.Print(result)
	case "monday_lessons":
		today := "Понедельник"
		result := find_day(today, sub, W)
		fmt.Print(result)
	case "tuesday_lessons":
		today := "Вторник"
		result := find_day(today, sub, W)
		fmt.Print(result)
	case "wednesday_lessons":
		today := "Среда"
		result := find_day(today, sub, W)
		fmt.Print(result)
	case "thursday_lessons":
		today := "Четверг"
		result := find_day(today, sub, W)
		fmt.Print(result)
	case "friday_lessons":
		today := "Пятница"
		result := find_day(today, sub, W)
		fmt.Print(result)
	case "com_to_lesson":
		num_of_lesson := 1
		LesNum := "lesson" + strconv.Itoa(num_of_lesson)
		comment := "Пары не будет"
		today := conv_day(time.Now().Weekday().String())
		filter := bson.D{{"day", today}, {"subgroup", sub}, {"week", W}}
		clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
		client, err := mongo.Connect(context.TODO(), clientOptions)
		var result Day
		collection := client.Database("schedule").Collection("PI-232")
		err = collection.FindOne(context.TODO(), filter).Decode(&result)
		upd := bson.D{{"$set", bson.D{{LesNum + ".comment", comment}}}}
		_, err = collection.UpdateOne(context.TODO(), filter, upd)
		if err != nil {
			log.Fatal(err)
		}
	case "where_group":
		today := conv_day(time.Now().Weekday().String())
		t := time.Now().String()
		t_h, _ := strconv.Atoi(t[11:13])
		t_m, _ := strconv.Atoi(t[14:16])
		result := find_day(today, sub, W)
		if t_h == 8 || (t_h == 9 && t_m < 50) {
			fmt.Print(result.Lesson1.Classroom)
		} else if (t_h == 9 && t_m >= 50) || (t_h == 10) || (t_h == 11 && t_m < 30) {
			fmt.Print(result.Lesson2.Classroom)
		} else if (t_h == 11 && t_m >= 30) || (t_h == 12) || (t_h == 13 && t_m < 20) {
			fmt.Print(result.Lesson3.Classroom)
		} else if (t_h == 13 && t_m >= 20) || (t_h == 14 && t_m < 50) {
			fmt.Print(result.Lesson4.Classroom)
		} else if (t_h == 15) || (t_h == 16 && t_m < 30) {
			fmt.Print(result.Lesson5.Classroom)
		}
	case "where_teacher":
		teacher := "Смирнова С. И."
		today := conv_day(time.Now().Weekday().String())
		t := time.Now().String()
		t_h, _ := strconv.Atoi(t[11:13])
		t_m, _ := strconv.Atoi(t[14:16])
		result := find_day(today, sub, W)
		if (t_h == 8 || (t_h == 9 && t_m < 50)) && (teacher == result.Lesson1.Teacher) {
			fmt.Print(result.Lesson1.Classroom)
		} else if ((t_h == 9 && t_m >= 50) || (t_h == 10) || (t_h == 11 && t_m < 30)) && (teacher == result.Lesson2.Teacher) {
			fmt.Print(result.Lesson2.Classroom)
		} else if ((t_h == 11 && t_m >= 30) || (t_h == 12) || (t_h == 13 && t_m < 20)) && (teacher == result.Lesson3.Teacher) {
			fmt.Print(result.Lesson3.Classroom)
		} else if ((t_h == 13 && t_m >= 20) || (t_h == 14 && t_m < 50)) && (teacher == result.Lesson4.Teacher) {
			fmt.Print(result.Lesson4.Classroom)
		} else if ((t_h == 15) || (t_h == 16 && t_m < 30)) && (teacher == result.Lesson5.Teacher) {
			fmt.Print(result.Lesson5.Classroom)
		}
	}
}

func main() {
	/*router := mux.NewRouter()                      //Создание роутера
	router.HandleFunc("/getSecret", handlerSecret) //Создание маршрута получения секретного кода
	router.HandleFunc("/getSchedule", getSchedule) //Создание маршрута получения JWT-токена
	http.ListenAndServe(":8082", nil)*/
	action := "next_lesson"
	sub := 1
	act(action, sub)

	/*day := Day{"Среда", "четная", 3, 1, Lessons{"Математика", "8:00", "лекция", "211А", "Смирнова С. И.", ""}, Lessons{"История", "9:50", "практика", "411В", "Дорофеев Д. В.", ""}, Lessons{"Математика", "11:30", "практика", "211А", "Смирнова С. И.", ""}, Lessons{}, Lessons{}}

	insertResult, err := collection.InsertOne(context.TODO(), day)
	if err != nil {
		log.Fatal(err)*/
}
