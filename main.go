package main

import (
	"Mattermost-bot-VK/bot"
	"Mattermost-bot-VK/config"
	"Mattermost-bot-VK/storage"
	"log"

	"github.com/mattermost/mattermost-server/v6/model"
)

func main() {
	tarcfg, matcfg := config.LoadConfig()
	st, err := storage.NewTarantoolStorage(tarcfg)
	if err != nil {
		log.Fatalf("Failed to initialize Tarantool storage: %v", err)
	}

	err = st.InitPollsSchema()
	if err != nil {
		log.Fatalf("Failed to create polls space: %v", err)
	}
	err = st.InitVotesSchema()
	if err != nil {
		log.Fatalf("Failed to create votes space: %v", err)
	}

	client := model.NewAPIv4Client(matcfg.URL)

	client.SetToken(config.MattermostToken)
	webSocketClient, err := model.NewWebSocketClient4(matcfg.WebSocketURL, client.AuthToken)
	if err != nil {
		log.Fatalf("WebSocket connection error: %v", err)
	}

	vote_bot := &bot.Bot{
		Client:          client,
		WebSocketClient: webSocketClient,
		Store:           st,
		UserID:          config.UserID,
		ChannelID:       config.ChannelID,
	}

	webSocketClient.Listen()
	go func() {
		for resp := range webSocketClient.EventChannel {
			vote_bot.HandleWebSocketResponse(resp)
		}
	}()

	log.Println("Bot started succsessfully")
	select {}
}
