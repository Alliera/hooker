package config

import (
	"gopkg.in/yaml.v2"
	"hooker/bot"
	"os"
)

type RepoConfig struct {
	Company            string        `yaml:"company"`
	WatchedBranches    []string      `yaml:"watched_branches"`
	RepoName           string        `yaml:"repo_name"`
	SlackAuthToken     string        `yaml:"slack_auth_token"`
	SlackAuthUserToken string        `yaml:"slack_auth_user_token"`
	SlackChannelID     string        `yaml:"slack_channel_id"`
	BotName            string        `yaml:"bot_name"`
	ShellCommand       string        `yaml:"shell_command"`
	CurrentBranch      string        `yaml:"-"`
	FinishBuild        func()        `yaml:"-"`
	Bot                *bot.SlackBot `yaml:"-"`
}

func LoadConfig(filename string) ([]*RepoConfig, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var configs []*RepoConfig
	err = yaml.Unmarshal(file, &configs)
	if err != nil {
		return nil, err
	}
	for _, conf := range configs {
		if conf.SlackChannelID != "" {
			conf.Bot = bot.NewSlackBot(conf.SlackAuthToken, conf.SlackAuthUserToken, conf.SlackChannelID, conf.BotName)
		}
	}

	return configs, nil
}
