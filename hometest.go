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
	"time"
)

const FILE_PATH string = "TheArtOfThinkingClearly.txt"
const SUMMARY_FILE_PATH string = "Summary.txt"
const CHAPTER_TITLE = "Summary of Chapter "

var askWG sync.WaitGroup
var appWg sync.WaitGroup

var originalContentQueue Queue
var readFileCompleted bool = false
var limit = 5 //concurrency up to n requests to OpenAI, to prevent hitting OpenAI's rate limiter

func main() {
	start := time.Now()
	fmt.Println("Start")

	appWg.Add(1)
	go readFile()

	appWg.Add(1)
	go summaryContentByAI()

	appWg.Wait()

	elapsed := time.Since(start) // End time
	fmt.Printf("Function took %s\n", elapsed)
}

func readFile() {
	defer appWg.Done()
	f, err := os.Open(FILE_PATH)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	var line string
	var content string
	for scanner.Scan() {
		line = scanner.Text()
		if isInteger(line) && len(content) > 0 {
			originalContentQueue.Enqueue(content)
			content = ""
		} else if !isInteger(line) {
			content += "\n" + line
		}
	}

	originalContentQueue.Enqueue(content)
	readFileCompleted = true
}

func summaryContentByAI() {
	defer appWg.Done()

	var index = 1
	for !readFileCompleted || !originalContentQueue.IsEmpty() {
		if len(originalContentQueue.items) > limit || readFileCompleted {
			var summaryContentArr = make([]string, limit)
			for i := range limit {
				var originalContent = originalContentQueue.Dequeue()
				askWG.Add(1)
				go askAI(originalContent, index, i, summaryContentArr)
				index++
			}

			askWG.Wait()

			go writeFile(summaryContentArr)
		}
	}
}

func askAI(content string, chaperIndex int, arrIndex int, summaryContentArr []string) {
	defer askWG.Done()
	key := "sk-team2024-home-test-vP00R3HgNtYAroFA1rRbT3BlbkFJ6XTugV2XnCDwPKOnwg18"

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
		var summaryContent = openAIResponse.Choices[0].Message.Content
		var title string = CHAPTER_TITLE + strconv.Itoa(chaperIndex)
		summaryContent = title + "\n" + summaryContent + "\n"
		summaryContentArr[arrIndex] = summaryContent
	}
}

func writeFile(summaryContentArr []string) {
	f, err := os.OpenFile(SUMMARY_FILE_PATH, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	check(err)
	defer f.Close()

	for index := range summaryContentArr {
		var content string = summaryContentArr[index]
		if len(content) > 0 {
			_, err := f.WriteString(content)
			check(err)
			f.Sync()
		}
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

type Queue struct {
	items []string
}

func (q *Queue) Enqueue(item string) {
	q.items = append(q.items, item)
}

func (q *Queue) Dequeue() string {
	if len(q.items) == 0 {
		fmt.Println("Queue is empty")
		return ""
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *Queue) IsEmpty() bool {
	return len(q.items) == 0
}

func (q *Queue) Front() string {
	if len(q.items) == 0 {
		fmt.Println("Queue is empty")
		return ""
	}
	return q.items[0]
}
