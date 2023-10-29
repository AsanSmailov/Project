package users

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type About struct {
	FirstName string `bson:"first_name"`
	LastName  string `bson:"last_name"`
	Group     string `bson:"group"`
}

type User struct {
	GithubID int64  `bson:"github_id"`
	TgId     int64  `bson:"tg_id"`
	Role     string `bson:"role"`
	About    About  `bson:"about"`
}

func checkData(id int64, idType string) bool {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	err = client.Connect(context.TODO())
	err = client.Ping(context.TODO(), nil)

	collection := client.Database("UsersDB").Collection("user")
	var field string

	if idType == "github" {
		field = "github_id"
	} else {
		field = "tg_id"
	}
	filter := bson.D{{field, id}}

	var result User

	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		err = client.Disconnect(context.TODO())
		return false
	} else {
		return true
	}
}

func register(userId int64, idType string) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	err = client.Connect(context.TODO())
	err = client.Ping(context.TODO(), nil)

	collection := client.Database("UsersDB").Collection("user")

	var user User
	if idType == "github" {
		user.GithubID = userId
	} else {
		user.TgId = userId
	}
	user.Role = "student"
	user.About.FirstName = ""
	user.About.LastName = ""
	user.About.Group = ""

	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Disconnect(context.TODO())
}
