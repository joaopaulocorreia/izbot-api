package main

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"

	waLog "go.mau.fi/whatsmeow/util/log"

	"context"
)

var WhatsappConnections = make(map[string]*Client)
var WhatsappManager *Whatsapp

type Whatsapp struct {
	db            *sqlstore.Container
	messageBroker *RabbitMQ

	dbLog     waLog.Logger
	clientLog waLog.Logger
}

func InitWhatsapp(messageBroker *RabbitMQ) *Whatsapp {
	if WhatsappManager != nil {
		return WhatsappManager
	}

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	clientLog := waLog.Stdout("Client", "DEBUG", true)

	db, err := sqlstore.New(context.Background(), "sqlite3", "file:whatsapp_store.db?_foreign_keys=on", dbLog)
	FailOnError(err, "[Whatsapp] Failed to open sqlite")

	WhatsappManager = &Whatsapp{
		db:            db,
		messageBroker: messageBroker,
		dbLog:         dbLog,
		clientLog:     clientLog,
	}

	return WhatsappManager
}

func (w Whatsapp) NewClient() *Client {
	store := w.db.NewDevice()

	client := &Client{
		WAClient:      whatsmeow.NewClient(store, w.clientLog),
		messageBroker: w.messageBroker,
	}
	client.register()

	return client
}

func (w Whatsapp) LoadClient(store *store.Device) *Client {
	client := &Client{
		WAClient:      whatsmeow.NewClient(store, w.clientLog),
		messageBroker: w.messageBroker,
	}
	client.register()

	return client
}

func (w Whatsapp) LoadAllDevices() {
	devicesStore, err := w.db.GetAllDevices(context.Background())
	FailOnError(err, "[SQLite] Failed to get all connected devices")

	for _, store := range devicesStore {
		client := w.LoadClient(store)
		err = client.Connect()
		FailOnError(err, "[Whatsapp] Error on connect client")
	}
}
