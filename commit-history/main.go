package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Commit struct {
	CommitID string `json:"commitId"`
	Comment  string `json:"comment"`
}

type Response struct {
	Value []Commit `json:"value"`
}

// This function fetches commits from Azure DevOps using the provided credentials and repository details.
func fetchCommits(author, user, token, org, project, repo string) ([]Commit, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/git/repositories/%s/commits?searchCriteria.author=%s&api-version=6.0", org, project, repo, author)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(user, token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Value, nil
}

// This function appends new commit entries to the specified file without overwriting existing content.
func appendToFile(filename string, commits []Commit) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, commit := range commits {
		_, err := file.WriteString(fmt.Sprintf("%s - %s\n", commit.CommitID, commit.Comment))
		if err != nil {
			return err
		}
	}
	return nil
}

func gitCommitAndPush(filename, message string) error {
	cmds := [][]string{
		{"git", "add", filename},
		{"git", "commit", "-m", message},
		{"git", "push"},
	}

	for _, cmd := range cmds {
		cmdExec := exec.Command(cmd[0], cmd[1:]...)
		cmdExec.Stdout = os.Stdout
		cmdExec.Stderr = os.Stderr
		if err := cmdExec.Run(); err != nil {
			return err
		}
	}

	return nil
}

func scheduleDailyTask() {
	for {
		now := time.Now()
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), 23, 0, 0, 0, now.Location())
		if now.After(nextRun) {
			nextRun = nextRun.Add(24 * time.Hour)
		}
		
		duration := nextRun.Sub(now)
		time.Sleep(duration)

		author := os.Getenv("AZURE_AUTHOR")
		user := os.Getenv("AZURE_USER")
		token := os.Getenv("AZURE_TOKEN")
		org := os.Getenv("AZURE_ORG")
		project := os.Getenv("AZURE_PROJECT")
		repo := os.Getenv("AZURE_REPO")
		filename := "commits.txt"

		fmt.Println("Buscando commits...")
		commits, err := fetchCommits(author, user, token, org, project, repo)
		if err != nil {
			fmt.Println("Erro ao buscar commits:", err)
			continue
		}

		fmt.Println("Adicionando commits ao arquivo...")
		if err := appendToFile(filename, commits); err != nil {
			fmt.Println("Erro ao escrever no arquivo:", err)
			continue
		}

		fmt.Println("Commitando e enviando para o GitHub...")
		if err := gitCommitAndPush(filename, "Atualizando commits do Azure DevOps"); err != nil {
			fmt.Println("Erro ao fazer commit e push:", err)
		}
	}
}

// Main function starts the scheduled task in a separate goroutine and keeps the program running indefinitely.
func main() {
	go scheduleDailyTask()
	select {}
}
