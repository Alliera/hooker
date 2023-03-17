package bot

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sfreiberg/progress"
	"github.com/slack-go/slack"
	"os"
	"strings"
	"time"
)

var token string
var channelID string
var userToken string
var clientUser *slack.Client
var bot *slack.Client
var ProgressMessage string
var lastExecutionTime time.Duration

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	token = os.Getenv("SLACK_AUTH_TOKEN")
	userToken = os.Getenv("SLACK_AUTH_USER_TOKEN")
	channelID = os.Getenv("SLACK_CHANNEL_ID")
	clientUser = slack.New(userToken, slack.OptionDebug(false))
	bot = slack.New(token, slack.OptionDebug(false))
	ProgressMessage = "Create build..."
	lastExecutionTime = 150 * time.Second
}
func NotifyBuildInfo(
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
	if tag != "" {
		blocks = append(
			blocks,
			createSection(
				"🏷️ "+"*Tag:*",
				"<https://github.com/Alliera/web-ui/releases/tag/"+tag+"|"+tag+">",
			))
	}
	blocks = append(
		blocks,
		createSection(
			"🔀 "+"*Branch:*",
			"<https://github.com/Alliera/web-ui/commits/"+branch+"|"+branch+">",
		))
	blocks = append(
		blocks,
		createSection(
			"📝 "+"*Message:*",
			"<https://github.com/Alliera/web-ui/commit/"+hash+"|"+commitMessage+">",
		))
	blocks = append(
		blocks,
		createSection(
			"#⃣️ "+"*Hash:*",
			"<https://github.com/Alliera/web-ui/commit/"+hash+"|"+hash[0:20]+">",
		))
	blocks = append(blocks, createSection("📆 "+"*Date:*", date))

	_, _, err := bot.PostMessage(
		channelID,
		slack.MsgOptionBlocks(blocks...),
	)

	if err != nil {
		fmt.Println(err)
	}
}
func NotifyFinished() {
	_, _, err := bot.PostMessage(
		channelID,
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
func ClearMessages(substring string) {
	params := slack.NewSearchParameters()
	var err error
	matchesCount := 1
	for err == nil && matchesCount > 0 {
		var messages *slack.SearchMessages
		query := "in:#web-ui-bot"
		if substring != "" {
			query += " " + substring
		}
		messages, err = clientUser.SearchMessages(query, params)
		matchesCount = len(messages.Matches)
		for _, message := range messages.Matches {
			_, _, err = bot.DeleteMessage(channelID, message.Timestamp)
			if err != nil {
				fmt.Println(err)
			}
		}
		params.Page += 1
	}
}
func Process(ctx context.Context) {
	startTime := time.Now()
	ClearMessages(ProgressMessage)
	opts := progress.DefaultOptions(ProgressMessage)
	opts.Width = 10
	opts.Fill = "🟥"
	opts.Empty = "⬛"
	pbar := progress.New(token, channelID, opts)
	opts.TotalUnits = int(lastExecutionTime / time.Second)
	i := 0
	for {
		select {
		case <-ctx.Done():
			_ = pbar.Update(opts.TotalUnits)
			timeDiff := time.Now().Sub(startTime)
			if timeDiff > 1*time.Minute {
				lastExecutionTime = timeDiff
			}
			return
		case <-time.After(1 * time.Second):
			if i < opts.TotalUnits {
				go func() {
					_ = pbar.Update(i)
				}()
			}
		}
		i++
	}
}
