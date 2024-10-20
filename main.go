package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/MumMumGoodBoy/restaurant-food-service/internal/service"
	"github.com/MumMumGoodBoy/restaurant-food-service/proto"
	"github.com/joho/godotenv"
	"github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file")
	}

	port := os.Getenv("PORT")
	// Connect to MongoDB
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://user:pass@localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, close := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		close()
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	fmt.Println("Connected to MongoDB")

	db := client.Database("restaurant-food-service")
	restaurantCollection := db.Collection("restaurant")
	foodCollection := db.Collection("food")

	// Connect to RabbitMQ
	rabbitMQConn, err := amqp091.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitMQConn.Close()

	rabbitMQChannel, err := rabbitMQConn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitMQChannel.Close()

	fmt.Println("Connected to RabbitMQ")

	if err = rabbitMQChannel.ExchangeDeclare(
		"restaurant_topic", // name
		"topic",            // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	); err != nil {
		log.Fatal(err)
	}

	if err = rabbitMQChannel.ExchangeDeclare(
		"food_topic", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	); err != nil {
		log.Fatal(err)
	}

	restaurantFoodService := service.RestaurantFoodService{
		RestaurantCollection: restaurantCollection,
		FoodCollection:       foodCollection,
		RabbitMQChannel:      rabbitMQChannel,
	}

	grpcServer := grpc.NewServer()
	proto.RegisterRestaurantFoodServer(grpcServer, &restaurantFoodService)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Restaurant Food Service is running on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
