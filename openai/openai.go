package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hometest/config"
	"io"
	"net/http"
	"strconv"
)

const CHAPTER_TITLE = "Summary of Chapter "
const PREFIX_QUESTION = "Summary \""
const ROLE_USER = "user"

func AskOpenAI(content string, chaperIndex int, arrIndex int, summaryContentArr []string) {
	key := config.GetConfig().OpenAIKey

	url := config.GetConfig().OpenAIURL
	question := PREFIX_QUESTION + content + "\""

	payload := map[string]interface{}{
		"model": config.GetConfig().OpenAIModel,
		"messages": []map[string]string{
			{"role": ROLE_USER, "content": question},
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
		summaryContentArr[arrIndex] = summaryContent // save summary content (response of OpenAI) to array
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
