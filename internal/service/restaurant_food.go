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
	restaurantId, err := bson.ObjectIDFromHex(food.RestaurantId)
	if err != nil {
		return nil, fmt.Errorf("invalid RestaurantId: %v", err)
	}
	foodModel := &model.Food{
		Id:           bson.NewObjectID(),
		RestaurantId: restaurantId,
		Name:         food.Name,
		Description:  food.Description,
		Price:        food.Price,
	}

	result, err := r.FoodCollection.InsertOne(ctx, foodModel)
	if err != nil {
		fmt.Println("Error inserting food: ", err)
		return nil, err
	}

	return &proto.Food{
		Id:           result.InsertedID.(bson.ObjectID).Hex(),
		RestaurantId: food.RestaurantId,
		Name:         food.Name,
		Description:  food.Description,
		Price:        food.Price,
	}, nil
}

// CreateRestaurant implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) CreateRestaurant(ctx context.Context, restaurant *proto.CreateRestaurantRequest) (*proto.Restaurant, error) {
	restaurantModel := &model.Restaurant{
		Id:      bson.NewObjectID(),
		Name:    restaurant.Name,
		Address: restaurant.Address,
		Phone:   restaurant.Phone,
	}
	result, err := r.RestaurantCollection.InsertOne(ctx, restaurantModel)

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
func (r *RestaurantFoodService) DeleteFood(ctx context.Context, food *proto.FoodIdRequest) (*proto.Empty, error) {
	foodId, err := bson.ObjectIDFromHex(food.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid FoodId: %v", err)
	}
	_, err = r.FoodCollection.DeleteOne(ctx, bson.M{"_id": foodId})
	if err != nil {
		fmt.Println("Error deleting food: ", err)
		return nil, err
	}
	return &proto.Empty{}, nil
}

// DeleteRestaurant implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) DeleteRestaurant(ctx context.Context, req *proto.RestaurantIdRequest) (*proto.Empty, error) {
	restaurantId, err := bson.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid RestaurantId: %v", err)
	}

	_, err = r.RestaurantCollection.DeleteOne(ctx, bson.M{"_id": restaurantId})
	if err != nil {
		fmt.Println("Error deleting restaurant:", err)
		return nil, err
	}

	return &proto.Empty{}, nil
}

// GetFoodByFoodId implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) GetFoodByFoodId(ctx context.Context, food *proto.FoodIdRequest) (*proto.Food, error) {
	foodId, err := bson.ObjectIDFromHex(food.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid FoodId: %v", err)
	}
	var foodModel model.Food
	err = r.FoodCollection.FindOne(ctx, bson.M{"_id": foodId}).Decode(&foodModel)
	if err != nil {
		fmt.Println("Error finding food: ", err)
		return nil, err
	}
	return &proto.Food{
		Id:           foodModel.Id.Hex(),
		RestaurantId: foodModel.RestaurantId.Hex(),
		Name:         foodModel.Name,
		Description:  foodModel.Description,
		Price:        foodModel.Price,
	}, nil
}

// GetFoodsByRestaurantId implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) GetFoodsByRestaurantId(ctx context.Context, restaurant *proto.RestaurantIdRequest) (*proto.GetFoodResponse, error) {
	restaurantId, err := bson.ObjectIDFromHex(restaurant.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid RestaurantId: %v", err)
	}
	cursor, err := r.FoodCollection.Find(ctx, bson.M{"restaurant_id": restaurantId})
	if err != nil {
		fmt.Println("Error finding food: ", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var foods []*proto.Food
	for cursor.Next(ctx) {
		var foodModel model.Food
		if err := cursor.Decode(&foodModel); err != nil {
			fmt.Println("Error decoding food: ", err)
			return nil, err
		}
		foods = append(foods, &proto.Food{
			Id:           foodModel.Id.Hex(),
			RestaurantId: foodModel.RestaurantId.Hex(),
			Name:         foodModel.Name,
			Description:  foodModel.Description,
			Price:        foodModel.Price,
		})
	}
	if err := cursor.Err(); err != nil {
		fmt.Println("Cursor error: ", err)
		return nil, err
	}
	return &proto.GetFoodResponse{Foods: foods}, nil
}

// GetRestaurantByRestaurantId implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) GetRestaurantByRestaurantId(ctx context.Context, req *proto.RestaurantIdRequest) (*proto.Restaurant, error) {
	restaurantId, err := bson.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid RestaurantId: %v", err)
	}

	result := r.RestaurantCollection.FindOne(ctx, bson.M{"_id": restaurantId})

	var restaurantModel model.Restaurant
	err = result.Decode(&restaurantModel)

	if err != nil {
		if err == mongo.ErrNoDocuments {
            return nil, fmt.Errorf("restaurant not found")
        }

        fmt.Errorf("Error finding restaurant: %v", err)
        return nil, err
	}

	restaurant := &proto.Restaurant {
		Id:      restaurantModel.Id.Hex(),
        Name:    restaurantModel.Name,
        Address: restaurantModel.Address,
        Phone:   restaurantModel.Phone,
	}

	return restaurant, nil
}

// GetRestaurants implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) GetRestaurants(ctx context.Context, _ *proto.Empty) (*proto.GetRestaurantResponse, error) {
	cursor, err := r.RestaurantCollection.Find(ctx, bson.M{})
	if err != nil {
		fmt.Println("Error finding restaurants:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var restaurants []*proto.Restaurant
	for cursor.Next(ctx) {
		var restaurantModel model.Restaurant
		if err := cursor.Decode(&restaurantModel); err != nil {
			fmt.Println("Error decoding restaurant:", err)
			return nil, err
		}
		restaurants = append(restaurants, &proto.Restaurant{
			Id:      restaurantModel.Id.Hex(),
			Name:    restaurantModel.Name,
			Address: restaurantModel.Address,
			Phone:   restaurantModel.Phone,
		})
	}
	if err := cursor.Err(); err != nil {
		fmt.Println("Cursor error:", err)
		return nil, err
	}

	return &proto.GetRestaurantResponse{Restaurants: restaurants}, nil
}

// UpdateFood implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) UpdateFood(ctx context.Context, food *proto.Food) (*proto.Food, error) {
	foodId, err := bson.ObjectIDFromHex(food.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid FoodId: %v", err)
	}
	restaurantId, err := bson.ObjectIDFromHex(food.RestaurantId)
	if err != nil {
		return nil, fmt.Errorf("invalid RestaurantId: %v", err)
	}

	foodModel := &model.Food{
		Id:           foodId,
		RestaurantId: restaurantId,
		Name:         food.Name,
		Description:  food.Description,
		Price:        food.Price,
	}
	_, err = r.FoodCollection.ReplaceOne(ctx, bson.M{"_id": foodId}, foodModel)
	if err != nil {
		fmt.Println("Error updating food: ", err)
		return nil, err
	}
	return &proto.Food{
		Id:           food.Id,
		RestaurantId: food.RestaurantId,
		Name:         food.Name,
		Description:  food.Description,
		Price:        food.Price,
	}, nil
}

// UpdateRestaurants implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) UpdateRestaurants(ctx context.Context, restaurant *proto.Restaurant) (*proto.Restaurant, error) {
	restaurantId, err := bson.ObjectIDFromHex(restaurant.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid RestaurantId: %v", err)
	}

	restaurantModel := &model.Restaurant{
		Id:      restaurantId,
		Name:    restaurant.Name,
		Address: restaurant.Address,
		Phone:   restaurant.Phone,
	}

	_, err = r.RestaurantCollection.ReplaceOne(ctx, bson.M{"_id": restaurantId}, restaurantModel)
	if err != nil {
		fmt.Println("Error updating restaurant:", err)
		return nil, err
	}

	return &proto.Restaurant{
		Id:      restaurant.Id,
		Name:    restaurant.Name,
		Address: restaurant.Address,
		Phone:   restaurant.Phone,
	}, nil
}
