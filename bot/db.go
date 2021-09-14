package bot

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func Connect() (*mongo.Client, error) {
	DB_URL := "mongodb+srv://mengere:89UJA330edsJC7EA@cluster0.3aigp.mongodb.net/Pbot?retryWrites=true&w=majority"
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(DB_URL))

	if err != nil {
		panic(err)
	}
	err = client.Ping(context.Background(), readpref.Primary())

	if err != nil {
		panic(err)
	}

	return client, nil
}

func GetAccount(client *mongo.Client, email string, result interface{}) error {
	collection := client.Database("Pbot").Collection("account")
	query := bson.M{"email": email}
	err := collection.FindOne(context.TODO(), query).Decode(result)
	return err
}
