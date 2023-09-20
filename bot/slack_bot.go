package bot

import (
	"context"
	"fmt"
	"github.com/sfreiberg/progress"
	"github.com/slack-go/slack"
	"strings"
	"time"
)

type SlackBot struct {
	ClientUser        *slack.Client
	Bot               *slack.Client
	ProgressMessage   string
	LastExecutionTime time.Duration
	ChannelID         string
	AuthToken         string
	BootName          string
}

func NewSlackBot(token string, userToken string, channelId string, botName string) *SlackBot {
	return &SlackBot{
		ClientUser:        slack.New(userToken, slack.OptionDebug(true)),
		Bot:               slack.New(channelId, slack.OptionDebug(false)),
		ProgressMessage:   "Create build...",
		LastExecutionTime: 150 * time.Second,
		ChannelID:         channelId,
		AuthToken:         token,
		BootName:          botName,
	}
}

func (s *SlackBot) NotifyBuildInfo(
	company string,
	repoName string,
	pusher string,
	author string,
	tag string,
	branch string,
	commitMessage string,
	date string,
	hash string,
) {
	stringSlice := strings.Split(date, "T")
	date = stringSlice[0] + " " + strings.Split(stringSlice[1], "+")[0]
	blocks := []slack.Block{
		slack.SectionBlock{
			Type: slack.MBTHeader,
			Text: &slack.TextBlockObject{
				Type:  "plain_text",
				Text:  "Starting build creation",
				Emoji: true,
			},
		},
	}
	blocks = append(blocks, slack.SectionBlock{
		Type: slack.MBTDivider,
	})
	blocks = append(blocks, createSection("🥷 "+"*Pusher:*", pusher))
	blocks = append(blocks, createSection("👨‍💻️ "+"*Author:*", author))
	baseUrl := "https://github.com/" + company + "/" + repoName
	if tag != "" {
		blocks = append(
			blocks,
			createSection(
				"🏷️ "+"*Tag:*",
				"<"+baseUrl+"/releases/tag/"+tag+"|"+tag+">",
			))
	}
	blocks = append(
		blocks,
		createSection(
			"🔀 "+"*Branch:*",
			"<"+baseUrl+"/commits/"+branch+"|"+branch+">",
		))
	blocks = append(
		blocks,
		createSection(
			"📝 "+"*Message:*",
			"<"+baseUrl+"/commit/"+hash+"|"+commitMessage+">",
		))
	blocks = append(
		blocks,
		createSection(
			"#⃣️ "+"*Hash:*",
			"<"+baseUrl+"/commit/"+hash+"|"+hash[0:20]+">",
		))
	blocks = append(blocks, createSection("📆 "+"*Date:*", date))

	_, _, err := s.Bot.PostMessage(
		s.ChannelID,
		slack.MsgOptionBlocks(blocks...),
	)

	if err != nil {
		fmt.Println(err)
	}
}
func (s *SlackBot) NotifyFinished() {
	_, _, err := s.Bot.PostMessage(
		s.ChannelID,
		slack.MsgOptionBlocks(slack.SectionBlock{
			Type: slack.MBTHeader,
			Text: &slack.TextBlockObject{
				Type:  "plain_text",
				Text:  "🎉Build Finished🎉",
				Emoji: true,
			},
		}, slack.SectionBlock{
			Type: slack.MBTDivider,
		}),
	)
	if err != nil {
		fmt.Println(err)
	}
}

func createSection(title string, text string) slack.SectionBlock {
	return slack.SectionBlock{
		Type: slack.MBTSection,
		Fields: []*slack.TextBlockObject{
			{
				Type: "mrkdwn",
				Text: title,
			},
			{
				Type: "mrkdwn",
				Text: text,
			},
		},
	}
}
func (s *SlackBot) ClearMessages(substring string) {
	params := slack.NewSearchParameters()
	var err error
	matchesCount := 1
	for err == nil && matchesCount > 0 {
		var messages *slack.SearchMessages
		query := "in:#" + s.BootName
		if substring != "" {
			query += " " + substring
		}
		resp, err := s.ClientUser.AuthTest()
		if err != nil {
			fmt.Println(resp)
			return
		}
		messages, err = s.ClientUser.SearchMessages(query, params)

		matchesCount = len(messages.Matches)
		for _, message := range messages.Matches {
			_, _, err = s.Bot.DeleteMessage(s.ChannelID, message.Timestamp)
			if err != nil {
				fmt.Println(err)
			}
		}
		params.Page += 1
	}
}
func (s *SlackBot) Process(ctx context.Context) {
	startTime := time.Now()
	s.ClearMessages(s.ProgressMessage)
	opts := progress.DefaultOptions(s.ProgressMessage)
	opts.Width = 10
	opts.Fill = "🟥"
	opts.Empty = "⬛"
	pbar := progress.New(s.AuthToken, s.ChannelID, opts)
	opts.TotalUnits = int(s.LastExecutionTime / time.Second)
	i := 0
	for {
		select {
		case <-ctx.Done():
			_ = pbar.Update(opts.TotalUnits)
			timeDiff := time.Now().Sub(startTime)
			if timeDiff > 1*time.Minute {
				s.LastExecutionTime = timeDiff
			}
			return
		case <-time.After(1 * time.Second):
			if i < opts.TotalUnits-(opts.TotalUnits/100)-1 {
				go func() {
					_ = pbar.Update(i)
				}()
			}
		}
		i++
	}
}
