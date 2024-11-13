package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

var commandQueue = &CommandQueue{}

func main() {
	eventHandler := dispatcher.NewEventDispatcher("", "").
		OnCustomizedEvent("im.message.receive_v1", HandleMessage)
	cli := larkws.NewClient(AppID, AppSecret,
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
	)

	ctx := context.Background()

	go processCommands(ctx, commandQueue)

	err := cli.Start(ctx)
	if err != nil {
		panic(err)
	}
}

func HandleMessage(ctx context.Context, event *larkevent.EventReq) error {
	var eventBody EventBody
	err := json.Unmarshal([]byte(event.Body), &eventBody)
	if err != nil {
		return err
	}

	message := eventBody.Event.Message.Content.Text

	// Start with a slash, it's a command
	if message[0] == '/' {
		parts := strings.Fields(eventBody.Event.Message.Content.Text)
		if len(parts) == 0 {
			return nil
		}

		command := parts[0]
		args := parts[1:]
		fmt.Println("Received command:", command, "Args:", args)
		commandQueue.Enqueue(Command{Type: command, Args: args, Event: eventBody.Event})
	} else {
		fmt.Println("Received message:", message)
		handleReply(ctx, Command{Type: message, Args: make([]string, 0), Event: eventBody.Event})
	}
	return nil
}
