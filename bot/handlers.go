package bot

import (
	"Mattermost-bot-VK/storage"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pborman/uuid"
)

func (b *Bot) SendMessage(message string) error {
	poll := &model.Post{
		ChannelId: b.ChannelID,
		Message:   message,
	}

	_, _, err := b.Client.CreatePost(poll)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bot) handleCreatePoll(commandArgs *model.CommandArgs) error {
	log.Println("Try to create poll")

	args := strings.SplitN(commandArgs.Command, "\"", -1)
	if len(args) < 5 {
		if err := b.SendMessage("Invalid request"); err != nil {
			return fmt.Errorf("error when sending a message: %e", err)
		}
		return nil
	}

	question := args[1]
	options := make(map[string]string)
	for i, opt := range args[3:] {
		if i%2 == 0 && len(opt) > 0 {
			optionID := fmt.Sprintf("opt%d", i/2+1)
			options[optionID] = opt
		}
	}

	if len(options) < 2 {
		if err := b.SendMessage("Poll must have at least 2 options"); err != nil {
			return fmt.Errorf("error when sending a message: %e", err)
		}
	}

	poll := storage.Poll{
		ID:        uuid.New(),
		CreatorID: commandArgs.UserId,
		Question:  question,
		Options:   options,
		CreatedAt: time.Now().Unix(),
		IsActive:  true,
	}

	pollID, err := b.Store.CreatePoll(poll)
	if err != nil {
		return fmt.Errorf("failed to create poll: %w", err)
	}

	var optionsText strings.Builder
	for id, text := range options {
		fmt.Fprintf(&optionsText, "%s: %s\n", id, text)
	}

	if err := b.SendMessage(fmt.Sprintf("Poll created succsessfully.\n To vote use command: /vote <pollID> <vote number>.\n PollID:%s", pollID)); err != nil {
		return fmt.Errorf("error when sending a message: %e", err)
	}

	log.Println("Created new poll")

	return nil
}

func (b *Bot) handleVote(commandArgs *model.CommandArgs) error {
	args := strings.Split(commandArgs.Command, " ")

	if len(args) < 4 {
		if err := b.SendMessage("Invalid request"); err != nil {
			return fmt.Errorf("error when sending a message: %e", err)
		}
		return nil
	}

	pollID := args[2]
	option := args[3]

	err := b.Store.Vote(pollID, commandArgs.UserId, option)
	if err == fmt.Errorf("poll does not exist") {
		if err := b.SendMessage("Poll does not exist"); err != nil {
			return fmt.Errorf("error when sending a message: %e", err)
		}
		return nil
	}
	if err != nil {
		if err := b.SendMessage("Something went wrong..."); err != nil {
			return fmt.Errorf("error when sending a message: %e", err)
		}
		return err
	}

	if err := b.SendMessage("Succsessfull voting"); err != nil {
		return fmt.Errorf("error when sending a message: %e", err)
	}

	return nil
}

func (b *Bot) handleResults(commandArgs *model.CommandArgs) error {
	args := strings.Split(commandArgs.Command, " ")

	if len(args) != 3 {
		if err := b.SendMessage("Invalid request"); err != nil {
			return fmt.Errorf("error when sending a message: %e", err)
		}
		return nil
	}

	pollID := args[2]

	results, err := b.Store.GetResults(pollID)
	if err == fmt.Errorf("poll does not exist") {
		if err := b.SendMessage("Poll does not exist"); err != nil {
			return fmt.Errorf("error when sending a message: %e", err)
		}
		return nil
	}
	if err != nil {
		if err := b.SendMessage("Something went wrong..."); err != nil {
			return fmt.Errorf("error when sending a message: %e", err)
		}
		return err
	}

	response, err := json.Marshal(results)
	if err != nil {
		if err := b.SendMessage("Succsessfull voting"); err != nil {
			return fmt.Errorf("error when sending a message: %e", err)
		}
		return err
	}

	if err := b.SendMessage("Results: " + string(response)); err != nil {
		return fmt.Errorf("error when sending a message: %e", err)
	}
	return err
}

func (b *Bot) handleEndPoll(commandArgs *model.CommandArgs) error {
	args := strings.Split(commandArgs.Command, " ")

	if len(args) != 3 {
		if err := b.SendMessage("Invalid request"); err != nil {
			return fmt.Errorf("error when sending a message: %e", err)
		}
		return nil
	}

	pollID := args[2]
	err := b.Store.EndPoll(pollID, commandArgs.UserId)
	if err == fmt.Errorf("poll does not exist") {
		if err := b.SendMessage("Poll does not exist"); err != nil {
			return fmt.Errorf("error when sending a message: %e", err)
		}
		return nil
	}
	if err != nil {
		if err := b.SendMessage("Something went wrong..."); err != nil {
			return fmt.Errorf("error when sending a message: %e", err)
		}
		return err
	}

	if err := b.SendMessage("End poll"); err != nil {
		return fmt.Errorf("error when sending a message: %e", err)
	}

	return nil
}

func (b *Bot) handleDeletePoll(commandArgs *model.CommandArgs) error {
	args := strings.Split(commandArgs.Command, " ")

	if len(args) != 3 {
		if err := b.SendMessage("Invalid request"); err != nil {
			return fmt.Errorf("error when sending a message: %e", err)
		}
		return nil
	}

	pollID := args[2]
	err := b.Store.DeletePoll(pollID, commandArgs.UserId)
	if err == fmt.Errorf("poll does not exist") {
		if err := b.SendMessage("Poll does not exist"); err != nil {
			return fmt.Errorf("error when sending a message: %e", err)
		}
		return nil
	}
	if err != nil {
		if err := b.SendMessage("Something went wrong..."); err != nil {
			return fmt.Errorf("error when sending a message: %e", err)
		}
		return err
	}

	if err := b.SendMessage("Delete poll"); err != nil {
		return fmt.Errorf("error when sending a message: %e", err)
	}

	return nil
}
