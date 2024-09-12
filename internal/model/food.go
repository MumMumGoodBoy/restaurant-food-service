package model

import "go.mongodb.org/mongo-driver/v2/bson"

type Food struct {
	Id           bson.ObjectID `bson:"_id"`
	RestaurantId bson.ObjectID `bson:"restaurant_id"`
	Name         string        `bson:"name"`
	Description  string        `bson:"description"`
	Price        float32       `bson:"price"`
}
