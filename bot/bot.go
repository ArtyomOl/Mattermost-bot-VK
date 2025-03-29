package bot

import (
	"Mattermost-bot-VK/config"
	"Mattermost-bot-VK/storage"
	"encoding/json"
	"log"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
)

type Bot struct {
	Client          *model.Client4
	WebSocketClient *model.WebSocketClient
	Store           *storage.TarantoolStorage
	Config          *config.MattermostConfig
	UserID          string
	ChannelID       string
}

func (b *Bot) HandleWebSocketResponse(event *model.WebSocketEvent) {
	if event.EventType() != model.WebsocketEventPosted {
		return
	}

	var post model.Post
	if err := json.Unmarshal([]byte(event.GetData()["post"].(string)), &post); err != nil {
		log.Printf("Unmarhall error: %v", err)
		return
	}

	if post.UserId == b.UserID {
		return
	}

	args := strings.Split(post.Message, "\"")
	if len(args) < 1 {
		return
	}
	commandArgs := &model.CommandArgs{UserId: b.UserID, ChannelId: b.ChannelID, Command: post.Message}
	if len(strings.Split(args[0], " ")) < 2 {
		return
	}
	if strings.Split(args[0], " ")[0] != "/poll" {
		return
	}
	method_type := strings.Split(args[0], " ")[1]

	switch method_type {
	case "create":
		if err := b.handleCreatePoll(commandArgs); err != nil {
			log.Printf("%e\n", err)
		}
	case "vote":
		if err := b.handleVote(commandArgs); err != nil {
			log.Printf("%e\n", err)
		}
	case "results":
		if err := b.handleResults(commandArgs); err != nil {
			log.Printf("%e\n", err)
		}
	case "end":
		if err := b.handleEndPoll(commandArgs); err != nil {
			log.Printf("%e\n", err)
		}
	case "delete":
		if err := b.handleDeletePoll(commandArgs); err != nil {
			log.Printf("%e\n", err)
		}
	default:
		return
	}
}
