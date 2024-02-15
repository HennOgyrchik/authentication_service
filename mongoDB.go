package main

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type connect struct {
	collection *mongo.Collection
	ctx        context.Context
}

type info struct {
	Token   string
	ExpTime string
}

func dbConn() *connect {

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://192.168.0.116:27017"))
	if err != nil {
		return nil
	}

	//проверка подключения
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil
	}

	collection := client.Database("tokens").Collection("refresh")
	if err != nil {
		return nil
	}

	return &connect{collection: collection, ctx: ctx}
}

func (c *connect) insertOne(token string, time time.Duration) (interface{}, error) {
	rows, err := c.find(token)
	if err != nil {
		return nil, err
	}

	if len(rows) != 0 {
		return nil, errors.New(alreadyExists)
	}

	insResult, err := c.collection.InsertOne(c.ctx, bson.M{"token": token, "expTime": time.String()})
	return insResult.InsertedID, err
}

func (c *connect) find(token string) ([]info, error) {
	var results []info

	cur, err := c.collection.Find(c.ctx, bson.M{"token": token})
	if err != nil {
		return nil, err
	}

	for cur.Next(c.ctx) {
		var element info
		err = cur.Decode(&element)
		if err != nil {
			return nil, err
		}
		results = append(results, element)

	}

	return results, nil
}
