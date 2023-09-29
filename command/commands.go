package command

import (
	"fmt"
	"hooker/bot"
	"hooker/config"
	"hooker/shell"
	"strings"
)

func HandleCommand(commandMessage string, config *config.ProjectConfig) {
	if !config.HasBot() {
		fmt.Println("Bot or commands is not configured")
		return
	}
	if config.HasCommands() {
		command := ""
		if strings.Contains(commandMessage, " ") {
			parts := strings.Split(commandMessage, " ")
			if len(parts) != 2 {
				sendMessage(config.Bot, "too many arguments")
				return
			}
			commandName := parts[0]
			commandArgument := clearArgument(parts[1])
			if !isCommand(config.Commands, commandName) {
				sendMessage(config.Bot, "invalid command")
				return
			}
			command = config.Commands[commandName]
			if !strings.Contains(command, "{}") {
				sendMessage(config.Bot, "command does not support argument")
				return
			}
			command = strings.ReplaceAll(command, "{}", commandArgument)
		} else {
			if isCommand(config.Commands, commandMessage) {
				command = config.Commands[commandMessage]
			}
		}
		if command != "" {
			ctx := config.GetContext()
			go func() {
				err := shell.Shellout(ctx, command, config.Bot)
				if err != nil {
					sendMessage(config.Bot, err.Error())
				}
			}()
			return
		}
	}
	switch commandMessage {
	case "stop":
		fmt.Println("Received stop commandMessage. Stopping...")
		if config.Stop != nil {
			config.Stop()
		}
	default:
		sendMessage(config.Bot, fmt.Sprintf("Received unknown commandMessage: %s\n", commandMessage))
	}
}

func clearArgument(text string) string {
	safeInput := strings.ReplaceAll(text, ";", "")
	safeInput = strings.ReplaceAll(safeInput, "&", "")
	return safeInput
}

func sendMessage(bot *bot.SlackBot, text string) {
	err := bot.SendPlainText(text)
	if err != nil {
		fmt.Println(err)
	}
}

func isCommand(commands map[string]string, text string) bool {
	for key := range commands {
		if strings.Contains(text, key) {
			return true
		}
	}
	return false
}
