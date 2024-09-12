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
