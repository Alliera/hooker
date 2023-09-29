package main

import (
	"fmt"
	"github.com/slack-go/slack"
	"hooker/bot"
	"hooker/command"
	"hooker/config"
	"hooker/git"
	"log"
	"reflect"
)

var configs []*config.ProjectConfig

func main() {
	var err error
	configs, err = config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	for _, conf := range configs {
		conf.Queue = make(chan interface{}, 100)
		if conf.HasBot() {
			conf.Bot = bot.NewSlackBot(
				conf.SlackAuthToken,
				conf.SlackAuthUserToken,
				conf.SlackChannelID,
				conf.BotName,
				conf.Queue,
			)
		}
		go startQueueHandler(conf.Queue, conf)
	}

	git.StartRestApiServer(configs)
}

func startQueueHandler(queue chan interface{}, config *config.ProjectConfig) {
	for message := range queue {
		switch castedMessage := message.(type) {
		case git.Body:
			git.ProcessGitHookBody(castedMessage, config)
		case slack.Message:
			// User message
			if castedMessage.BotID == "" {
				command.HandleCommand(castedMessage.Text, config)
			}
		default:
			fmt.Printf("Received a non-Body message: %v\n", reflect.TypeOf(message))
		}
	}
}
