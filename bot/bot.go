package bot

import (
	"Mattermost-bot-VK/config"
	"Mattermost-bot-VK/storage"
	"fmt"
	"log"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
)

type VoteBot struct {
	client *model.Client4
	store  storage.Storage
	config *config.MattermostConfig
	userID string
}

func NewVoteBot(client *model.Client4, store storage.Storage, cfg *config.MattermostConfig) *VoteBot {
	return &VoteBot{
		client: client,
		store:  store,
		config: cfg,
	}
}

func (b *VoteBot) Start() error {
	// Получаем информацию о боте
	user, _, err := b.client.GetMe("")
	if err != nil {
		return fmt.Errorf("failed to get bot user: %w", err)
	}
	b.userID = user.Id

	// Регистрируем команды
	if err := b.registerCommands(); err != nil {
		return fmt.Errorf("failed to register commands: %w", err)
	}

	log.Println("Bot started successfully")
	return nil
}

func (b *VoteBot) Stop() {
	log.Println("Bot stopped")
}

func (b *VoteBot) registerCommands() error {
	command := &model.Command{
		Trigger:          "vote",
		AutoComplete:     true,
		AutoCompleteDesc: "Manage polls: create, vote, results, end, delete",
		AutoCompleteHint: "[command] [args]",
		DisplayName:      "Vote Bot",
		Description:      "Create and manage polls in channels",
		URL:              fmt.Sprintf("%s/plugins/com.mattermost.vote-bot/command", b.config.URL),
	}

	_, _, err := b.client.CreateCommand(command)
	return err
}

func (b *VoteBot) HandleCommand(commandArgs *model.CommandArgs) (*model.CommandResponse, error) {
	args := strings.Fields(commandArgs.Command)

	switch args[1] {
	case "create":
		return b.handleCreatePoll(commandArgs)
	case "vote":
		return b.handleVote(commandArgs)
	case "results":
		return b.handleResults(commandArgs)
	case "end":
		return b.handleEndPoll(commandArgs)
	case "delete":
		return b.handleDeletePoll(commandArgs)
	default:
		return nil, fmt.Errorf("invalid request")
	}
}
