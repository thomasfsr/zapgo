package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func eventHandler(client *whatsmeow.Client) func(interface{}) {
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			message := v.Message.GetConversation()
			sender := v.Info.Sender

			fmt.Println("Received a message!", message)

			if message != "" {
				// response := fmt.Sprintf("You said: %s", message)
				// res, _ := callLLM(message)
				// fmt.Println(res)

				groq_res, error1 := callGroq(message)

				fmt.Print("\n", error1)

				fmt.Println(groq_res)
				groq_res_str := string(groq_res.Function.Arguments)

				_, err := client.SendMessage(context.Background(), sender, &waE2E.Message{
					Conversation: &groq_res_str,
				})

				if err != nil {
					fmt.Printf("Error sending message: %v\n", err)
				} else {
					fmt.Println("Response sent successfully") // Added newline
				}
			}
		}
	}
}

func main() {
	// |------------------------------------------------------------------------------------------------------|
	// | NOTE: You must also import the appropriate DB connector, e.g. github.com/mattn/go-sqlite3 for SQLite |
	// |------------------------------------------------------------------------------------------------------|

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	dbPath := os.Getenv("DB_PATH")
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on", dbPath)
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "sqlite3", dsn, dbLog)
	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler(client))

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}
