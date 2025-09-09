package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mdp/qrterminal"

	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type Client struct {
	WAClient       *whatsmeow.Client
	eventHandlerID uint32
	ID             string

	messageBroker *RabbitMQ
}

func (c Client) Connect() error {
	err := c.WAClient.Connect()

	return err
}

func (c *Client) register() {
	c.eventHandlerID = c.WAClient.AddEventHandler(c.eventHandler)
}

func (c *Client) eventHandler(evt any) {
	switch v := evt.(type) {
	case *events.Connected:
		c.ID = c.WAClient.Store.ID.User
		Connections[c.ID] = c

		queueName := fmt.Sprintf("client-%s", c.ID)
		dataChan := c.messageBroker.Consume(queueName)

		go func() {
			for data := range dataChan {
				fmt.Printf("[RabbitMQ] Received a message: %s \n", data.Body)

				var message Message
				err := json.Unmarshal([]byte(data.Body), &message)
				if err != nil {
					fmt.Println(err)
					continue
				}

				time.Sleep(10 * time.Second)

				targetID := types.NewJID(message.To, types.DefaultUserServer)
				c.WAClient.SendMessage(context.Background(), targetID, &waE2E.Message{
					Conversation: proto.String(message.Text),
				})
			}
		}()

		fmt.Printf("[whatsapp][%s] New client connected\n", c.ID)
	case *events.Message:
		fmt.Printf("[whatsapp][%s] New Message received: %s\n", c.ID, v.RawMessage.GetConversation())
	case *events.QR:
		fmt.Println("[whatsapp] New QRCODE generated")
		qrterminal.GenerateHalfBlock(v.Codes[0], qrterminal.L, os.Stdout)
	case *events.LoggedOut:
		delete(Connections, c.ID)

		fmt.Printf("[whatsapp][%s] Logout\n", c.ID)
	case *events.Disconnected:
		delete(Connections, c.ID)

		fmt.Printf("[whatsapp][%s] Disconnected\n", c.ID)
	}
}
