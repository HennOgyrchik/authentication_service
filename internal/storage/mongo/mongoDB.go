package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Connect struct {
	collection *mongo.Collection
	ctx        context.Context
}

type data struct {
	Token   string
	ExpTime int64
	Guid    string
}

// DBConn Создание подключения к БД
func DBConn(addr string, port string) (*Connect, error) {
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

	return &Connect{collection: collection, ctx: ctx}, err
}

// InsertOne Вставка одной записи в БД
func (c *Connect) InsertOne(guid string, token string, expTime int64) error {
	_, err := c.collection.InsertOne(c.ctx, bson.M{"token": token, "expTime": expTime, "guid": guid})
	return err
}

// Find поиск всех записей в БД по guid
func (c *Connect) Find(guid string) ([]data, error) {
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

// FindOne поиск одной записи по токену
func (c *Connect) FindOne(token string) (data, error) {
	var result data
	err := c.collection.FindOne(c.ctx, bson.M{"token": token}).Decode(&result)
	return result, err

}

// DeleteOne удаление одной записи из БД по хэшу
func (c *Connect) DeleteOne(hash string) error {
	row, err := c.FindOne(hash)
	if err != nil {
		return err
	}

	_, err = c.collection.DeleteOne(c.ctx, bson.M{"token": row.Token})
	return err
}

// Drop очистка БД
func (c *Connect) Drop() error {
	return c.collection.Drop(c.ctx)
}
