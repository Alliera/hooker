package git

import (
	"encoding/json"
	"fmt"
	"hooker/config"
	"hooker/helper"
	"io"
	"log"
	"net/http"
	"time"
)

func StartRestApiServer(configs []*config.ProjectConfig) {
	handler := http.NewServeMux()
	handler.HandleFunc("/getHook", Logger(hook(configs)))
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

func hook(configs []*config.ProjectConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		for _, repoConfig := range configs {
			if hookBody.Repository.Name == repoConfig.RepoName {
				if branch == repoConfig.CurrentBranch {
					repoConfig.Stop()
					time.Sleep(1 * time.Second)
				}
				if repoConfig.WatchedBranches == nil ||
					(len(repoConfig.WatchedBranches) != 0 && helper.Contains(repoConfig.WatchedBranches, branch)) {
					repoConfig.Queue <- hookBody
				}
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log.Printf("Server [http] method [%s] connnection from [%v]", r.Method, r.RemoteAddr)
		next.ServeHTTP(w, r)
	}
}
