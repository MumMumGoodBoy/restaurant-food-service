package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/MumMumGoodBoy/restaurant-food-service/internal/service"
	"github.com/MumMumGoodBoy/restaurant-food-service/proto"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://user:pass@localhost:27017"))

	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	fmt.Println("Connected")

	db := client.Database("restaurant-food-service")
	restaurantCollection := db.Collection("restaurant")
	foodCollection := db.Collection("food")

	restaurantService := service.RestaurantFoodService{
		RestaurantCollection: restaurantCollection,
		FoodCollection:       foodCollection,
	}
	grpcServer := grpc.NewServer()
	proto.RegisterRestaurantFoodServer(grpcServer, &restaurantService)

	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
