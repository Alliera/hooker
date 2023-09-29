package git

import (
	"context"
	"fmt"
	"hooker/config"
	"hooker/shell"
	"strings"
)

func ProcessGitHookBody(body Body, projectConfig *config.ProjectConfig) {
	fmt.Println("Start Processing: ", body.Ref)
	if body.Ref == "" {
		fmt.Println("Branch is not defined", body.Ref)
		return
	}
	ref := strings.Replace(body.Ref, "refs/heads/", "", -1)

	var tag string
	tag, projectConfig.CurrentBranch = getTagAndBranch(body)
	ctx := projectConfig.GetContext()
	if projectConfig.Bot != nil {
		projectConfig.Bot.NotifyBuildInfo(
			projectConfig.Company,
			projectConfig.RepoName,
			body.Pusher.Name,
			body.HeadCommit.Author.Name,
			tag,
			projectConfig.CurrentBranch,
			body.HeadCommit.Message,
			body.HeadCommit.Timestamp,
			body.HeadCommit.Id,
		)

		go projectConfig.Bot.Process(ctx)
	}
	updateGit(projectConfig.HomeFolder, ref, projectConfig.RepoName)
	_ = shell.Shellout(ctx, projectConfig.ShellCommand)

	if projectConfig.Stop != nil {
		projectConfig.Stop()
	}

	if projectConfig.Bot != nil {
		projectConfig.Bot.NotifyFinished()
	}

}

func getTagAndBranch(body Body) (tag string, branch string) {
	branch = strings.Replace(body.Ref, "refs/heads/", "", -1)
	if body.BaseRef != nil {
		branch = strings.Replace(*body.BaseRef, "refs/heads/", "", -1)
		tag = strings.Replace(body.Ref, "refs/tags/", "", -1)
	}
	return
}

func updateGit(homeFolder string, branch string, projectFolderName string) {
	cmd := "cd " + homeFolder + "/" + projectFolderName + " && " +
		"git reset --hard && git checkout master && git pull && " +
		"for b in `git branch --merged | grep -v \\*`; do git branch -D $b; done && " +
		"git checkout " + branch + " && " +
		"git pull"
	fmt.Println(cmd)
	err := shell.Shellout(context.Background(), cmd)
	if err != nil {
		fmt.Println("Failed to update git:", err)
	}
}
