package dao

import (
	"context"
	"fmt"

	"github.com/jonathanface/storm-reporter/API/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StormDAO struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewStormDAO(uri, dbName, collName string) (*StormDAO, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	collection := client.Database(dbName).Collection(collName)
	return &StormDAO{client: client, collection: collection}, nil
}

func (dao *StormDAO) Disconnect() error {
	return dao.client.Disconnect(context.TODO())
}

func (dao *StormDAO) GetStormReports(startOfDay, endOfDay int64) ([]models.StormReport, error) {
	filter := bson.M{
		"date": bson.M{
			"$gte": startOfDay,
			"$lte": endOfDay,
		},
	}

	cursor, err := dao.collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query MongoDB: %w", err)
	}
	defer cursor.Close(context.TODO())

	var reports []models.StormReport
	if err := cursor.All(context.TODO(), &reports); err != nil {
		return nil, fmt.Errorf("failed to decode storm reports: %w", err)
	}

	return reports, nil
}
