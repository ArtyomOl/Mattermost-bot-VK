package storage

import (
	"Mattermost-bot-VK/config"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/viciious/go-tarantool"
)

type TarantoolStorage struct {
	conn *tarantool.Connection
}

func NewTarantoolStorage(cfg config.TarantoolConfig) (*TarantoolStorage, error) {
	opts := tarantool.Options{
		User: cfg.User,
	}

	conn, err := tarantool.Connect(cfg.Address, &opts)
	if err != nil {
		return nil, err
	}

	return &TarantoolStorage{conn: conn}, nil
}

func (s *TarantoolStorage) InitPollsSchema() error {
	query := &tarantool.Call{Name: "box.schema.create_space", Tuple: []interface{}{
		"polls",
		map[string]interface{}{
			"format": []map[string]interface{}{
				{"name": "id", "type": "string"},
				{"name": "creator_id", "type": "string"},
				{"name": "question", "type": "string"},
				{"name": "options", "type": "map"},
				{"name": "created_at", "type": "unsigned"},
				{"name": "is_active", "type": "boolean"},
			},
		},
	}}
	s.conn.Exec(context.Background(), query)

	query = &tarantool.Call{
		Name: "box.space.polls:create_index",
		Tuple: []interface{}{
			"primary",
			map[string]interface{}{
				"type":          "hash",
				"parts":         []interface{}{"id"},
				"if_not_exists": true,
			},
		},
	}
	resp := s.conn.Exec(context.Background(), query)
	if resp.Error != nil {
		return fmt.Errorf("failed to create space index: %w", resp.Error)
	}

	log.Println("Init space")

	return nil
}

func (s *TarantoolStorage) InitVotesSchema() error {
	query := &tarantool.Call{Name: "box.schema.create_space", Tuple: []interface{}{
		"votes",
		map[string]interface{}{
			"format": []map[string]interface{}{
				{"name": "id", "type": "string"},
				{"name": "creator_id", "type": "string"},
				{"name": "option", "type": "string"},
				{"name": "created_at", "type": "unsigned"},
			},
		},
	}}
	s.conn.Exec(context.Background(), query)

	query = &tarantool.Call{
		Name: "box.space.votes:create_index",
		Tuple: []interface{}{
			"primary",
			map[string]interface{}{
				"type":          "hash",
				"parts":         []interface{}{"id"},
				"if_not_exists": true,
			},
		},
	}
	resp := s.conn.Exec(context.Background(), query)
	if resp.Error != nil {
		return fmt.Errorf("failed to create space index: %w", resp.Error)
	}

	log.Println("Init space")

	return nil
}

func (s *TarantoolStorage) DeleteVotesSpace() error {
	// Lua-скрипт для проверки существования space и его удаления
	luaCode := `
        if box.space.votes ~= nil then
            box.space.votes:drop()
            return true
        else
            return false, "space 'votes' does not exist"
        end
    `

	query := &tarantool.Eval{
		Expression: luaCode,
	}

	resp := s.conn.Exec(context.Background(), query)

	if resp.Error != nil {
		return fmt.Errorf("failed to execute Lua script: %v", resp.Error)
	}

	if len(resp.Data) == 0 {
		return fmt.Errorf("no response from Tarantool")
	}

	return nil

}

func (s *TarantoolStorage) CreatePoll(poll Poll) (string, error) {
	options := make(map[string]interface{})
	for k, v := range poll.Options {
		options[k] = v
	}

	query := &tarantool.Insert{Space: "polls", Tuple: []interface{}{
		poll.ID,
		poll.CreatorID,
		poll.Question,
		options,
		uint64(poll.CreatedAt),
		poll.IsActive,
	},
	}
	result := s.conn.Exec(context.Background(), query)
	if result.Error != nil {
		return "", result.Error
	}

	log.Println("Create poll")

	return poll.ID, nil
}

func (s *TarantoolStorage) GetPoll(id string) (*Poll, error) {
	query := &tarantool.Select{Space: "polls", Index: "primary", KeyTuple: []interface{}{id}}
	resp := s.conn.Exec(context.Background(), query)

	if resp.Error != nil {
		return nil, resp.Error
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("poll not found")
	}
	data := resp.Data[0]
	options := make(map[string]string)
	if opts, ok := data[3].(map[interface{}]interface{}); ok {
		for k, v := range opts {
			if key, ok := k.(string); ok {
				if value, ok := v.(string); ok {
					options[key] = value
				}
			}
		}
	}

	createdAt := time.Unix(int64(data[4].(uint64)), 0)
	isActive := data[5].(bool)

	log.Println("Get poll")

	return &Poll{
		ID:        data[0].(string),
		CreatorID: data[1].(string),
		Question:  data[2].(string),
		Options:   options,
		CreatedAt: createdAt.Unix(),
		IsActive:  isActive,
	}, nil
}

func (s *TarantoolStorage) Vote(pollID, userID, option string) error {
	poll, err := s.GetPoll(pollID)
	if err != nil {
		return err
	}
	if !poll.IsActive {
		return fmt.Errorf("poll is not active")
	}

	query := &tarantool.Insert{Space: "votes", Tuple: []interface{}{
		pollID,
		userID,
		option,
		uint64(time.Now().Unix()),
	},
	}
	resp := s.conn.Exec(context.Background(), query)
	if resp.Error != nil {
		log.Println(resp.Error)
	}

	log.Println("Succsessfull vote")

	return nil
}

func (s *TarantoolStorage) GetResults(pollID string) (*PollResult, error) {
	poll, err := s.GetPoll(pollID)
	if err != nil {
		return nil, fmt.Errorf("failed to get poll: %v", err)
	}

	query := &tarantool.Select{
		Space:    "votes",
		Index:    "primary",
		KeyTuple: []interface{}{pollID},
	}
	resp := s.conn.Exec(context.Background(), query)

	if resp.Error != nil {
		return nil, fmt.Errorf("failed to get votes: %v", resp.Error)
	}

	votes := make(map[string]int)
	totalVotes := 0

	for _, item := range resp.Data {
		if option, ok := item[2].(string); ok {
			votes[option]++
			totalVotes++
		}
	}

	results := &PollResult{
		PollID:     poll.ID,
		Question:   poll.Question,
		Options:    poll.Options,
		Votes:      votes,
		TotalVotes: totalVotes,
		IsActive:   poll.IsActive,
	}

	log.Println("Get result")

	return results, nil
}

func (s *TarantoolStorage) EndPoll(pollID, userID string) error {
	poll, err := s.GetPoll(pollID)
	if err != nil {
		return err
	}
	if poll.CreatorID != userID {
		return fmt.Errorf("only poll creator can end the poll")
	}

	query := &tarantool.Update{Space: "polls", Index: "primary", Key: pollID, KeyTuple: []interface{}{
		[]interface{}{"=", 5, false},
	}}
	resp := s.conn.Exec(context.Background(), query)

	if resp.Error != nil {
		fmt.Errorf("poll does not exist")
	}

	log.Println("End poll")

	return nil
}

func (s *TarantoolStorage) DeletePoll(pollID, userID string) error {
	poll, err := s.GetPoll(pollID)
	if err != nil {
		return err
	}
	if poll.CreatorID != userID {
		return fmt.Errorf("only poll creator can delete the poll")
	}

	query := &tarantool.Delete{Space: "polls", KeyTuple: []interface{}{pollID}}
	resp := s.conn.Exec(context.Background(), query)
	if resp.Error != nil {
		return fmt.Errorf("poll does not exist")
	}

	query = &tarantool.Delete{Space: "votes", KeyTuple: []interface{}{pollID}}
	resp = s.conn.Exec(context.Background(), query)
	if resp.Error != nil {
		return resp.Error
	}

	log.Println("Delete poll")

	return nil
}

func (s *TarantoolStorage) Close() {
	s.conn.Close()
}
