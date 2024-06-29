package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

const FILE_PATH string = "TheArtOfThinkingClearly.txt"
const CHAPTER_TITLE = "Summary of Chapter "

var summaryContentArr [20]string
var wg sync.WaitGroup

type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func main() {
	readFileAndAskOpenAi()

	wg.Wait()

	writeFile()
}

func readFileAndAskOpenAi() {
	f, err := os.Open(FILE_PATH)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	var line string
	var content string
	var index int64
	for scanner.Scan() {
		// do something with a line
		line = scanner.Text()
		if isInteger(line) && len(content) > 0 {
			index, _ = strconv.ParseInt(line, 0, 64)
			wg.Add(1)
			go summaryContentByOpenAi(index-2, content)
			content = ""
		} else if !isInteger(line) {
			content += "\n" + line
		}
	}

	wg.Add(1)
	go summaryContentByOpenAi(index-1, content)
}

func writeFile() {
	f, err := os.Create("Summary.txt")
	check(err)
	defer f.Close()

	for index := range summaryContentArr {
		var content string = summaryContentArr[index]
		if len(content) > 0 {
			var title string = CHAPTER_TITLE + strconv.Itoa(index+1)
			content = title + content + "\n"
			_, err := f.WriteString(content)
			check(err)
			f.Sync()
		}
	}
}

func summaryContentByOpenAi(index int64, content string) {
	key := "key"

	url := "https://api.openai.com/v1/chat/completions"
	question := "Summary \"" + content + "\""

	payload := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": question},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshalling payload:", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making API request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var openAIResponse OpenAIResponse
	err = json.Unmarshal(body, &openAIResponse)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	if len(openAIResponse.Choices) > 0 {
		summaryContentArr[index] = openAIResponse.Choices[0].Message.Content
	}
}

func isInteger(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
