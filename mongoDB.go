package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type connect struct {
	collection *mongo.Collection
	ctx        context.Context
}

type data struct {
	Token   string
	ExpTime int64
	Guid    string
}

func dbConn() *connect {

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://192.168.0.116:27017"))
	if err != nil {
		return nil
	}

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

func (c *connect) insertOne(data data) error {
	_, err := c.collection.InsertOne(c.ctx, bson.M{"token": data.Token, "expTime": data.ExpTime, "guid": data.Guid})
	return err
}

func (c *connect) find(guid string) ([]data, error) {
	var result []data

	cur, err := c.collection.Find(c.ctx, bson.M{"guid": guid})
	if err != nil {
		return result, err
	}

	for cur.Next(c.ctx) {
		var element data
		err = cur.Decode(&element)
		if err != nil {
			return result, err
		}
		result = append(result, element)
	}

	return result, nil
}

func (c *connect) findOne(token string) (data, error) {
	var result data
	err := c.collection.FindOne(c.ctx, bson.M{"token": token}).Decode(&result)
	return result, err

}

func (c *connect) deleteOne(hash string) error {
	row, err := c.findOne(hash)
	if err != nil {
		return err
	}

	_, err = c.collection.DeleteOne(c.ctx, bson.M{"token": row.Token})
	return err
}
