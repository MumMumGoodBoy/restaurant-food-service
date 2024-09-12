package service

import (
	"context"
	"fmt"

	"github.com/MumMumGoodBoy/restaurant-food-service/internal/model"
	"github.com/MumMumGoodBoy/restaurant-food-service/proto"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var _ proto.RestaurantFoodServer = (*RestaurantFoodService)(nil)

type RestaurantFoodService struct {
	proto.UnimplementedRestaurantFoodServer
	RestaurantCollection *mongo.Collection
	FoodCollection       *mongo.Collection
}

// CreateFood implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) CreateFood(ctx context.Context, food *proto.Food) (*proto.Food, error) {
	panic("unimplemented")
}

// CreateRestaurant implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) CreateRestaurant(ctx context.Context, restaurant *proto.CreateRestaurantRequest) (*proto.Restaurant, error) {
	foodModel := &model.Restaurant{
		Id:      bson.NewObjectID(),
		Name:    restaurant.Name,
		Address: restaurant.Address,
		Phone:   restaurant.Phone,
	}
	result, err := r.RestaurantCollection.InsertOne(ctx, foodModel)

	if err != nil {
		fmt.Println("Error inserting restaurant: ", err)
		return nil, err
	}

	return &proto.Restaurant{
		Id:      result.InsertedID.(bson.ObjectID).Hex(),
		Name:    restaurant.Name,
		Address: restaurant.Address,
		Phone:   restaurant.Phone,
	}, nil
}

// DeleteFood implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) DeleteFood(context.Context, *proto.FoodIdRequest) (*proto.Empty, error) {
	panic("unimplemented")
}

// DeleteRestaurant implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) DeleteRestaurant(context.Context, *proto.RestaurantIdRequest) (*proto.Empty, error) {
	panic("unimplemented")
}

// GetFoodByFoodId implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) GetFoodByFoodId(context.Context, *proto.FoodIdRequest) (*proto.Food, error) {
	panic("unimplemented")
}

// GetFoodByRestaurantId implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) GetFoodByRestaurantId(context.Context, *proto.RestaurantIdRequest) (*proto.GetFoodResponse, error) {
	panic("unimplemented")
}

// GetRestaurantByRestaurantId implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) GetRestaurantByRestaurantId(context.Context, *proto.RestaurantIdRequest) (*proto.Restaurant, error) {
	panic("unimplemented")
}

// GetRestaurants implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) GetRestaurants(context.Context, *proto.Empty) (*proto.GetRestaurantResponse, error) {
	panic("unimplemented")
}

// UpdateFood implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) UpdateFood(context.Context, *proto.Food) (*proto.Food, error) {
	panic("unimplemented")
}

// UpdateRestaurants implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) UpdateRestaurants(context.Context, *proto.Restaurant) (*proto.Restaurant, error) {
	panic("unimplemented")
}
