package storage

import (
	"Mattermost-bot-VK/config"
	"context"
	"fmt"
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

func (s *TarantoolStorage) InitSchema(schemaName string) error {
	query := &tarantool.Call{Name: "box.space.create", Tuple: []interface{}{
		schemaName,
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

	query = &tarantool.Call{Name: "box.space." + schemaName + ":create_index", Tuple: []interface{}{"primary", map[string]interface{}{"parts": []string{"id"}, "if_not_exists": true}}}

	resp := s.conn.Exec(context.Background(), query)
	if resp.Error != nil {
		return fmt.Errorf("failed to create polls space index: %w", resp.Error)
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
		poll.CreatedAt.Unix(),
		poll.IsActive,
	},
	}
	resp := s.conn.Exec(context.Background(), query)

	if len(resp.Data) == 0 {
		return "", fmt.Errorf("no data returned from Tarantool")
	}

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

	createdAt := time.Unix(data[4].(int64), 0)
	isActive := data[5].(bool)

	return &Poll{
		ID:        data[0].(string),
		CreatorID: data[1].(string),
		Question:  data[2].(string),
		Options:   options,
		CreatedAt: createdAt,
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
	_, exists := poll.Options[option]
	if !exists {
		return fmt.Errorf("invalid option")
	}

	query := &tarantool.Insert{Space: "votes", Tuple: []interface{}{
		pollID,
		userID,
		option,
		time.Now().Unix(),
	},
	}
	s.conn.Exec(context.Background(), query)

	return nil
}

func (s *TarantoolStorage) GetResults(pollID string) (map[string]int, error) {
	poll, err := s.GetPoll(pollID)
	if err != nil {
		return nil, err
	}

	results := make(map[string]int)
	for opt := range poll.Options {
		results[opt] = 0
	}

	query := &tarantool.Select{Space: "polls", Index: "primary", KeyTuple: []interface{}{pollID}}
	resp := s.conn.Exec(context.Background(), query)
	if resp.Error != nil {
		return nil, resp.Error
	}

	for _, item := range resp.Data {
		data := item
		option := data[2].(string)
		results[option]++
	}

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

	query := &tarantool.Update{Space: "polls", Index: "primary", Key: []interface{}{pollID}, KeyTuple: []interface{}{
		[]interface{}{"=", 5, false},
	}}
	resp := s.conn.Exec(context.Background(), query)

	if resp.Error != nil {
		return resp.Error
	}

	return err
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
		return resp.Error
	}

	query = &tarantool.Delete{Space: "votes", KeyTuple: []interface{}{pollID}}
	resp = s.conn.Exec(context.Background(), query)
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (s *TarantoolStorage) Close() {
	s.conn.Close()
}
