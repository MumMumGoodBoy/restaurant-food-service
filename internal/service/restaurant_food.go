package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/MumMumGoodBoy/restaurant-food-service/internal/model"
	"github.com/MumMumGoodBoy/restaurant-food-service/proto"
	"github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var _ proto.RestaurantFoodServer = (*RestaurantFoodService)(nil)

type RestaurantFoodService struct {
	proto.UnimplementedRestaurantFoodServer
	RestaurantCollection *mongo.Collection
	FoodCollection       *mongo.Collection
	RabbitMQChannel      *amqp091.Channel
}

func (r *RestaurantFoodService) publishRestaurantEvent(data model.RestaurantEvent, event string) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling restaurant event: %v", err)
	}

	err = r.RabbitMQChannel.Publish(
		"restaurant_topic",
		event,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("error publishing restaurant event: %v", err)
	}
	return nil
}

func (r *RestaurantFoodService) publishFoodEvent(data model.FoodEvent, event string) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling food event: %v", err)
	}

	err = r.RabbitMQChannel.Publish(
		"food_topic",
		event,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("error publishing food event: %v", err)
	}
	return nil
}

// CreateFood implements prot                                o.RestaurantFoodServer.
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
		ImageUrl:     food.ImageUrl,
	}

	result, err := r.FoodCollection.InsertOne(ctx, foodModel)
	if err != nil {
		fmt.Println("Error inserting food: ", err)
		return nil, err
	}

	foodProto := &proto.Food{
		Id:           result.InsertedID.(bson.ObjectID).Hex(),
		RestaurantId: food.RestaurantId,
		Name:         food.Name,
		Description:  food.Description,
		Price:        food.Price,
		ImageUrl:     food.ImageUrl,
	}

	data := model.FoodEvent{
		Event:        "food.create",
		Id:           foodProto.Id,
		FoodName:     foodProto.Name,
		RestaurantId: foodProto.RestaurantId,
		Price:        foodProto.Price,
		Description:  foodProto.Description,
		ImageUrl:     foodProto.ImageUrl,
	}
	if err := r.publishFoodEvent(data, "food.create"); err != nil {
		fmt.Println("Error publishing food event: ", err)
	}

	return foodProto, nil
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
	restaurantProto := &proto.Restaurant{
		Id:      result.InsertedID.(bson.ObjectID).Hex(),
		Name:    restaurant.Name,
		Address: restaurant.Address,
		Phone:   restaurant.Phone,
	}

	event := model.RestaurantEvent{
		Event:          "restaurant.create",
		Id:             restaurantProto.Id,
		RestaurantName: restaurantProto.Name,
		Address:        restaurantProto.Address,
		Phone:          restaurantProto.Phone,
	}
	if err := r.publishRestaurantEvent(event, "restaurant.create"); err != nil {
		fmt.Println("Error publishing restaurant event: ", err)
	}

	return restaurantProto, nil
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
	data := model.FoodEvent{
		Event: "food.delete",
		Id:    food.Id,
	}
	if err := r.publishFoodEvent(data, "food.delete"); err != nil {
		fmt.Println("Error publishing food event: ", err)
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
	data := model.RestaurantEvent{
		Event: "restaurant.delete",
		Id:    req.Id,
	}
	if err := r.publishRestaurantEvent(data, "restaurant.delete"); err != nil {
		fmt.Println("Error publishing restaurant event: ", err)
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
		ImageUrl:     foodModel.ImageUrl,
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
			ImageUrl:     foodModel.ImageUrl,
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

		return nil, fmt.Errorf("error finding restaurant: %v", err)
	}

	restaurant := &proto.Restaurant{
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
		ImageUrl:     food.ImageUrl,
	}
	_, err = r.FoodCollection.ReplaceOne(ctx, bson.M{"_id": foodId}, foodModel)
	if err != nil {
		fmt.Println("Error updating food: ", err)
		return nil, err
	}

	data := model.FoodEvent{
		Event:        "food.update",
		Id:           food.Id,
		FoodName:     food.Name,
		RestaurantId: food.RestaurantId,
		Price:        food.Price,
		Description:  food.Description,
		ImageUrl:     food.ImageUrl,
	}

	if err := r.publishFoodEvent(data, "food.update"); err != nil {
		fmt.Println("Error publishing food event: ", err)
	}

	return &proto.Food{
		Id:           food.Id,
		RestaurantId: food.RestaurantId,
		Name:         food.Name,
		Description:  food.Description,
		Price:        food.Price,
		ImageUrl:     food.ImageUrl,
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

	event := model.RestaurantEvent{
		Event:          "restaurant.update",
		Id:             restaurant.Id,
		RestaurantName: restaurant.Name,
		Address:        restaurant.Address,
		Phone:          restaurant.Phone,
	}
	if err := r.publishRestaurantEvent(event, "restaurant.update"); err != nil {
		fmt.Println("Error publishing restaurant event: ", err)
	}

	return &proto.Restaurant{
		Id:      restaurant.Id,
		Name:    restaurant.Name,
		Address: restaurant.Address,
		Phone:   restaurant.Phone,
	}, nil
}

// GetFoodsByFoodIds implements proto.RestaurantFoodServer.
func (r *RestaurantFoodService) GetFoodsByFoodIds(ctx context.Context, req *proto.FoodIdsRequest) (*proto.GetFoodResponse, error) {
	var ids []bson.ObjectID
	for _, id := range req.Ids {
		oid, err := bson.ObjectIDFromHex(id)
		if err != nil {
			return nil, fmt.Errorf("invalid FoodId: %v", err)
		}
		ids = append(ids, oid)
	}

	cursor, err := r.FoodCollection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		slog.WarnContext(ctx, "Error finding foods by ids: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var foods []*proto.Food
	for cursor.Next(ctx) {
		var foodModel model.Food
		if err := cursor.Decode(&foodModel); err != nil {
			slog.WarnContext(ctx, "Error decoding food: %v", err)
			return nil, err
		}
		foods = append(foods, &proto.Food{
			Id:           foodModel.Id.Hex(),
			RestaurantId: foodModel.RestaurantId.Hex(),
			Name:         foodModel.Name,
			Description:  foodModel.Description,
			Price:        foodModel.Price,
			ImageUrl:     foodModel.ImageUrl,
		})
	}

	if err := cursor.Err(); err != nil {
		slog.WarnContext(ctx, "Cursor error: %v", err)
		return nil, err
	}

	return &proto.GetFoodResponse{Foods: foods}, nil
}
