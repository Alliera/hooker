package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hooker/bot"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const ShellToUse = "bash"

var queue chan Body

var currentWebUiBranch string

var finishWebUiBuild func()

func main() {
	queue = make(chan Body, 300)
	go startQueueHandler()
	startRestApiServer()
}

type Body struct {
	Ref        string     `json:"ref"`
	BaseRef    *string    `json:"base_ref"`
	After      string     `json:"after"`
	HeadCommit HeadCommit `json:"head_commit"`
	Repository Repository `json:"repository"`
	Pusher     User       `json:"pusher"`
}

type Repository struct {
	Name string `json:"name"`
}

type HeadCommit struct {
	Id        string `json:"id"`
	Author    User   `json:"author"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func startQueueHandler() {
	for body := range queue {
		fmt.Println("Start Processing: ", body.Ref)
		if body.Ref == "" {
			fmt.Println("Branch is not defined", body.Ref)
			continue
		}
		ref := strings.Replace(body.Ref, "refs/heads/", "", -1)
		Shellout(context.Background(), "source /var/www/hooker/.env")
		if body.Repository.Name == "web-ui" {
			var tag string
			tag, currentWebUiBranch = getTagAndBranch(body)
			bot.NotifyBuildInfo(
				body.Pusher.Name,
				body.HeadCommit.Author.Name,
				tag,
				currentWebUiBranch,
				body.HeadCommit.Message,
				body.HeadCommit.Timestamp,
				body.HeadCommit.Id,
			)
			var ctx context.Context
			ctx, finishWebUiBuild = context.WithCancel(context.Background())
			go bot.Process(ctx)
			updateGit(ref, "web-ui")
			Shellout(ctx, "cd /var/www/hooker/ && docker-compose build web_ui")
			finishWebUiBuild()
			bot.NotifyFinished()
			currentWebUiBranch = ""
		} else if body.Repository.Name == "xircl-api" {
			updateGit(ref, "xircl-api")
			Shellout(context.Background(), "cd /var/www/hooker/ && docker-compose up sourceguardian")
		} else {
			fmt.Println("Not target commit, skip...")
		}
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

func updateGit(branch string, projectFolderName string) {
	cmd := "cd /var/www/hooker/" + projectFolderName + " && " +
		"git reset --hard && git checkout master && git pull && " +
		"for b in `git branch --merged | grep -v \\*`; do git branch -D $b; done && " +
		"git checkout " + branch + " && " +
		"git pull"
	fmt.Println(cmd)
	Shellout(context.Background(), cmd)
}

func startRestApiServer() {
	handler := http.NewServeMux()
	handler.HandleFunc("/getHook", Logger(hook))
	port := ":7479"
	s := http.Server{
		Addr:           "0.0.0.0" + port,
		Handler:        handler,
		ReadTimeout:    0,
		WriteTimeout:   0,
		IdleTimeout:    0,
		MaxHeaderBytes: 1 << 20, //1*2^20 - 128 kByte
	}
	fmt.Println("REST server started on " + port)
	log.Println(s.ListenAndServe())
}

func hook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println("================================================")
	fmt.Println(string(body))
	fmt.Println("================================================")
	var hookBody Body
	err = json.Unmarshal(body, &hookBody)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, branch := getTagAndBranch(hookBody)
	if branch == currentWebUiBranch {
		finishWebUiBuild()
		time.Sleep(1 * time.Second)
	}
	queue <- hookBody
	w.WriteHeader(http.StatusOK)
}

func Shellout(ctx context.Context, command string) {
	cmd := exec.CommandContext(ctx, ShellToUse, "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	_ = cmd.Run()
}

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log.Printf("Server [http] method [%s] connnection from [%v]", r.Method, r.RemoteAddr)
		next.ServeHTTP(w, r)
	}
}
