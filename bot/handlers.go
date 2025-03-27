package bot

import (
	"Mattermost-bot-VK/storage"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mattermost/mattermost-server/v6/model"
)

func (b *VoteBot) handleCreatePoll(commandArgs *model.CommandArgs) (*model.CommandResponse, error) {
	args := strings.SplitN(commandArgs.Command, "\"", -1)
	if len(args) < 5 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Invalid format",
			ChannelId:    commandArgs.ChannelId,
		}, nil
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
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Poll must have at least 2 options",
			ChannelId:    commandArgs.ChannelId,
		}, nil
	}

	poll := storage.Poll{
		ID:        uuid.New().String(),
		CreatorID: commandArgs.UserId,
		Question:  question,
		Options:   options,
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	pollID, err := b.store.CreatePoll(poll)
	if err != nil {
		return nil, fmt.Errorf("failed to create poll: %w", err)
	}

	var optionsText strings.Builder
	for id, text := range options {
		fmt.Fprintf(&optionsText, "%s: %s\n", id, text)
	}

	response := fmt.Sprintf(
		"Poll created successfully!\nID: `%s`\nQuestion: %s\nOptions:\n%s\nTo vote use: `/vote %s option_id`",
		pollID, question, optionsText.String(), pollID,
	)

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeInChannel,
		Text:         response,
		ChannelId:    commandArgs.ChannelId,
	}, nil
}

func (b *VoteBot) handleVote(commandArgs *model.CommandArgs) (*model.CommandResponse, error) {
	args := strings.Fields(commandArgs.Command)

	pollID := args[2]
	option := args[3]

	err := b.store.Vote(pollID, commandArgs.UserId, option)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Error voting: %v", err),
			ChannelId:    commandArgs.ChannelId,
		}, nil
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "Vote recorded",
		ChannelId:    commandArgs.ChannelId,
	}, nil
}

func (b *VoteBot) handleResults(commandArgs *model.CommandArgs) (*model.CommandResponse, error) {
	args := strings.Split(commandArgs.Command, "\"")
	pollID := args[1]

	results, err := b.store.GetResults(pollID)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Error: %v", err),
			ChannelId:    commandArgs.ChannelId,
		}, nil
	}

	response, err := json.Marshal(results)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Error: %v", err),
			ChannelId:    commandArgs.ChannelId,
		}, nil
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeInChannel,
		Text:         string(response[:]),
		ChannelId:    commandArgs.ChannelId,
	}, nil
}
func (b *VoteBot) handleEndPoll(commandArgs *model.CommandArgs) (*model.CommandResponse, error) {
	args := strings.Split(commandArgs.Command, "\"")
	pollID := args[1]
	err := b.store.EndPoll(pollID, commandArgs.UserId)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Error: %v", err),
			ChannelId:    commandArgs.ChannelId,
		}, nil
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "Voting has been completed successfully",
		ChannelId:    commandArgs.ChannelId,
	}, nil
}

func (b *VoteBot) handleDeletePoll(commandArgs *model.CommandArgs) (*model.CommandResponse, error) {
	args := strings.Split(commandArgs.Command, "\"")
	pollID := args[1]
	err := b.store.DeletePoll(pollID, commandArgs.UserId)
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Error: %v", err),
			ChannelId:    commandArgs.ChannelId,
		}, nil
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "Voting has been deleted successfully",
		ChannelId:    commandArgs.ChannelId,
	}, nil
}
