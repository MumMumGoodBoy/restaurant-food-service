package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Restaurant struct {
	Id      bson.ObjectID `bson:"_id"`
	Name    string        `bson:"name"`
	Address string        `bson:"address"`
	Phone   string        `bson:"phone"`
}

type RestaurantEvent struct {
	Event          string `json:"event"`
	Id             string `json:"id"`
	RestaurantName string `json:"restaurantName"`
	Address        string `json:"address"`
	Phone          string `json:"phone"`
}
