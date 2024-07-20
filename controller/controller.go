package controller

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	dbName         string = "Image_Effect"
	collectionName string = "Image_Collection"
	uri            string = "mongodb+srv://ItishKiit:ItishKiit@cluster0.mxcf61e.mongodb.net/?retryWrites=true&w=majority"
)

var (
	client     *mongo.Client
	Collection *mongo.Collection
	err        error
)

func init() {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	clientOption := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err = mongo.Connect(context.TODO(), clientOption)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB")
	Collection = client.Database(dbName).Collection(collectionName)
}
