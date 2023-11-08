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

func checkData(githubID int64, tgID int64) bool {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	err = client.Connect(context.TODO())
	err = client.Ping(context.TODO(), nil)

	collection := client.Database("UsersDB").Collection("user")

	filter := bson.D{{"github_id", githubID}, {"tg_id", tgID}}
	log.Print(tgID, " ", githubID)
	var result User

	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		err = client.Disconnect(context.TODO())
		return false
	} else {
		return true
	}
}

func register(githubID int64, tgID int64) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	err = client.Connect(context.TODO())
	err = client.Ping(context.TODO(), nil)

	collection := client.Database("UsersDB").Collection("user")
	var user User
	log.Print(tgID, " ", githubID)
	user.GithubID = githubID
	user.TgId = tgID
	user.Role = "student"
	user.About.FullName = ""
	user.About.Group = ""

	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Disconnect(context.TODO())
}
