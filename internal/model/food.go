package model

import "go.mongodb.org/mongo-driver/v2/bson"

type Food struct {
	Id           bson.ObjectID `bson:"_id"`
	RestaurantId bson.ObjectID `bson:"restaurant_id"`
	Name         string        `bson:"name"`
	Description  string        `bson:"description"`
	Price        float32       `bson:"price"`
	ImageUrl	 string		   `bson:"image_url"`
}

type FoodEvent struct {
	Event        string  `json:"event"`
	Id           string  `json:"id"`
	FoodName     string  `json:"foodName"`
	RestaurantId string  `json:"restaurantId"`
	Description	 string	 `json:"description"`
	Price        float32 `json:"price"`
	ImageUrl	 string	 `json:"image_url"`
}
