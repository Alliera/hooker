package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	//      "time"
)

const ShellToUse = "bash"

var queue chan Body

func main() {
	queue = make(chan Body, 300)
	go startQueueHandler()
	startRestApiServer()
}

type Body struct {
	Ref        string     `json:"ref"`
	After      string     `json:"after"`
	HeadCommit HeadCommit `json:"head_commit"`
}

type HeadCommit struct {
	Added    []string `json:"added"`
	Removed  []string `json:"removed"`
	Modified []string `json:"modified"`
}

func startQueueHandler() {
	for body := range queue {
		fmt.Println("Start Processing: ", body.Ref)
		if body.Ref == "" {
			fmt.Println("Branch is not defined", body.Ref)
			continue
		}
		branch := strings.Replace(body.Ref, "refs/heads/", "", -1)
		hc := body.HeadCommit
		if commitHasWord(hc, "/react/") && commitHasWord(hc, "/Xircl/") {
			updateGit(branch)
			Shellout("cd /var/www/hooker/ && docker-compose build xircl_react && docker-compose up sourceguardian")
		} else if commitHasWord(hc, "/react/") {
			updateGit(branch)
			Shellout("cd /var/www/hooker/ && docker-compose build xircl_react")
		} else if commitHasWord(hc, "/Xircl/") {
			updateGit(branch)
			Shellout("cd /var/www/hooker/ && docker-compose up sourceguardian")
		} else {
			fmt.Println("Not target commit, skip...")
		}
	}
}

func commitHasWord(hc HeadCommit, keyWord string) bool {
	return hasWord(hc.Added, keyWord) || hasWord(hc.Modified, keyWord) || hasWord(hc.Removed, keyWord)
}

func updateGit(branch string) {
	cmd := "cd /var/www/hooker/xircl && " +
		"git reset --hard && git checkout develop && git pull && " +
		"git branch | grep -v \"develop\" | grep -v \"master\" | xargs git branch -D && " +
		"git checkout " + branch + " && " +
		"git pull"
	Shellout(cmd)
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
	body, err := ioutil.ReadAll(r.Body)
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
	queue <- hookBody
	w.WriteHeader(http.StatusOK)
}

func Shellout(command string) {
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
}

func hasWord(list []string, word string) bool {
	for _, v := range list {
		if strings.Contains(v, word) {
			return true
		}
	}
	return false
}

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log.Printf("Server [http] method [%s] connnection from [%v]", r.Method, r.RemoteAddr)
		next.ServeHTTP(w, r)
	}
}
