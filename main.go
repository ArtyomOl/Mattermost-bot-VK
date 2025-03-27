package main

import (
	"Mattermost-bot-VK/config"
	"Mattermost-bot-VK/storage"
	"fmt"
	"log"
	"time"
)

func main() {
	tarcfg := config.TarantoolConfig{Address: "127.0.0.1:3301"}
	s, err := storage.NewTarantoolStorage(tarcfg)
	if err != nil {
		log.Fatalf("Failed to initialize Tarantool storage: %v", err)
	}

	err = s.InitSchema("polls")
	if err != nil {
		panic(err)
	}
	err = s.InitSchema("votes")
	if err != nil {
		panic(err)
	}

	new_poll := storage.Poll{ID: "1234",
		CreatorID: "123",
		Question:  "Test_question",
		Options:   map[string]string{"u12": "test1", "u23": "test2"},
		CreatedAt: time.Now(),
		IsActive:  true}
	s.CreatePoll(new_poll)

	poll, err := s.GetPoll("1234")
	if err != nil {
		panic(err)
	}

	fmt.Println(*poll)

	defer s.Close()
}
