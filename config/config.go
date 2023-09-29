package config

import (
	"context"
	"gopkg.in/yaml.v2"
	"hooker/bot"
	"os"
)

type ProjectConfig struct {
	Company            string            `yaml:"company"`
	WatchedBranches    []string          `yaml:"watched_branches"`
	Commands           map[string]string `yaml:"commands"`
	HomeFolder         string            `yaml:"home_folder"`
	RepoName           string            `yaml:"repo_name"`
	SlackAuthToken     string            `yaml:"slack_auth_token"`
	SlackAuthUserToken string            `yaml:"slack_auth_user_token"`
	SlackChannelID     string            `yaml:"slack_channel_id"`
	BotName            string            `yaml:"bot_name"`
	ShellCommand       string            `yaml:"shell_command"`
	Queue              chan interface{}  `yaml:"-"`
	CurrentBranch      string            `yaml:"-"`
	Context            context.Context   `yaml:"-"`
	Stop               func()            `yaml:"-"`
	Bot                *bot.SlackBot     `yaml:"-"`
}

func (c *ProjectConfig) HasBot() bool {
	return c.SlackChannelID != ""
}
func (c *ProjectConfig) HasCommands() bool {
	return c.Commands != nil
}

func (c *ProjectConfig) GetContext() context.Context {
	if c.Context != nil {
		select {
		case <-c.Context.Done():
		default:
			return c.Context
		}
	}
	c.Context, c.Stop = context.WithCancel(context.Background())
	return c.Context
}

func LoadConfig(filename string) ([]*ProjectConfig, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var configs []*ProjectConfig
	err = yaml.Unmarshal(file, &configs)
	if err != nil {
		return nil, err
	}
	for _, conf := range configs {
		if conf.HomeFolder == "" {
			panic("home_folder is not defined")
		}
	}

	return configs, nil
}
