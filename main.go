package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	err := godotenv.Load()
	FailOnError(err, "[ENV] Error on load env")
	fmt.Println("[ENV] Envs loaded with success")

	message_broker_connection_string := os.Getenv("MESSAGE_BROKER_CONNECTION_STRING")

	messageBroker, err := NewRabbitMQ(message_broker_connection_string)
	FailOnError(err, "[MessageBroker] Error on connect to message broker")
	fmt.Println("[MessageBroker] Connected")

	wm := InitWhatsapp(messageBroker)
	wm.LoadAllDevices()
}
