package users

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type About struct {
	FullName string `bson:"full_name"`
	Group    string `bson:"group"`
}

type User struct {
	GithubID int64  `bson:"github_id"`
	TgId     int64  `bson:"tg_id"`
	Role     string `bson:"role"`
	About    About  `bson:"about"`
}

// Функция, проверяет наличие документа с github и tg id
func checkData(githubID int64, tgID int64) bool {
	//Соединяемся с MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	err = client.Connect(context.TODO())
	err = client.Ping(context.TODO(), nil)
	//Подключаемся к коллекции user
	collection := client.Database("UsersDB").Collection("user")
	//Формируем фильтр поиска
	filter := bson.D{{"github_id", githubID}, {"tg_id", tgID}}
	log.Print(tgID, " ", githubID)
	var result User
	//Ищем документ, если документа нет, то возвращает false
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		err = client.Disconnect(context.TODO())
		return false
	} else {
		err = client.Disconnect(context.TODO())
		return true
	}
}

// Функция, создания нового документа
func register(githubID int64, tgID int64) {
	//Соединяемся с MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	err = client.Connect(context.TODO())
	err = client.Ping(context.TODO(), nil)
	//Подключаемся к коллекции user
	collection := client.Database("UsersDB").Collection("user")

	//Формируем пользователя
	var user User
	log.Print(tgID, " ", githubID)
	user.GithubID = githubID
	user.TgId = tgID
	user.Role = "student"
	user.About.FullName = ""
	user.About.Group = ""
	//Создаём новый документ
	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Fatal(err)
	}
	//Отключаемся от MongoDB
	err = client.Disconnect(context.TODO())
}

// Функция, по id возвращает данные пользователя
func getData(ID int64, IDtype string) User {
	//Соединяемся с MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	err = client.Connect(context.TODO())
	err = client.Ping(context.TODO(), nil)
	//Подключаемся к коллекции user
	collection := client.Database("UsersDB").Collection("user")
	//Формируем фильтр поиска
	filter := bson.D{{IDtype, ID}}

	var result User
	//Получаем данные
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Disconnect(context.TODO())
	return result
}

// Функция, изменяет данные в документы, на вход принимает tg id, данные, тип данных, возвращает bool
func inputData(tgID int64, data string, datatype string) bool {
	//Соединяемся с MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	err = client.Connect(context.TODO())
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	//Подключаемся к коллекции user
	collection := client.Database("UsersDB").Collection("user")
	//Формируем фильтр поиска
	filter := bson.D{{"tg_id", tgID}}
	//Формируем фильтр данных, которые будем изменять
	var replacement bson.D
	if datatype != "role" {
		replacement = bson.D{{"$set", bson.D{{"about." + datatype, data}}}}
	} else {
		replacement = bson.D{{datatype, data}}
	}
	//Изменяем данные
	_, err1 := collection.UpdateOne(context.TODO(), filter, replacement)
	if err1 != nil {
		err1 = client.Disconnect(context.TODO())
		return false
	} else {
		err = client.Disconnect(context.TODO())
		return true
	}
}
