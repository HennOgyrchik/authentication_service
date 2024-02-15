package main

import (
	"context"
	"fmt"
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

// dbConn Создание подключения к БД
func dbConn(addr string, port string) (*connect, error) {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", addr, port)))
	if err != nil {
		return nil, err
	}

	//проверка соединения
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	collection := client.Database("tokens").Collection("refresh")

	return &connect{collection: collection, ctx: ctx}, err
}

// insertOne Вставка одной записи в БД
func (c *connect) insertOne(data data) error {
	_, err := c.collection.InsertOne(c.ctx, bson.M{"token": data.Token, "expTime": data.ExpTime, "guid": data.Guid})
	return err
}

// find поиск всех записей в БД по guid
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

// findOne поиск одной записи по токену
func (c *connect) findOne(token string) (data, error) {
	var result data
	err := c.collection.FindOne(c.ctx, bson.M{"token": token}).Decode(&result)
	return result, err

}

// deleteOne удаление одной записи из БД по хэшу
func (c *connect) deleteOne(hash string) error {
	row, err := c.findOne(hash)
	if err != nil {
		return err
	}

	_, err = c.collection.DeleteOne(c.ctx, bson.M{"token": row.Token})
	return err
}
