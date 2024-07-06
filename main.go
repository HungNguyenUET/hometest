package main

import (
	"bufio"
	"fmt"
	"hometest/config"
	"hometest/openai"
	"hometest/util"
	"log"
	"os"
	"sync"
	"time"
)

var FilePath string = config.GetConfig().FilePath
var SummaryFilePath string = config.GetConfig().SummaryFilePath
var OpenAiLimitRequest = config.GetConfig().OpenAiLimitRequest //concurrency up to n requests to OpenAI, to prevent hitting OpenAI's rate limiter

var askWG sync.WaitGroup // wait group for limit concurrent request send to OpenAI api
var appWg sync.WaitGroup // wait group for application read file and summary content by AI

var originalContentQueue util.Queue // Divide book to chapters and send to queue
var readFileCompleted bool = false  // Check when read file function read file completely

func main() {
	start := time.Now() // Monitor application execute time

	appWg.Add(1)
	go readFileAndDivideIntoChapters() // Run read file concurrently

	appWg.Add(1)
	go summaryContent() // summary each chapter by send to OpenAI concurrently

	appWg.Wait() // Wait application execute completely

	elapsed := time.Since(start) // End time
	fmt.Printf("Function took %s\n", elapsed)
}

/*
Read file by line, if line is chapter number then divide to chapter
Send content of each chapter and send to queue
*/
func readFileAndDivideIntoChapters() {
	defer appWg.Done()

	f, err := os.Open(FilePath)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	var line string
	var content string
	for scanner.Scan() {
		line = scanner.Text()
		if util.IsInteger(line) && len(content) > 0 { // Check if line is chapter number
			originalContentQueue.Enqueue(content) // Send current chapter to queue and scan next chapter
			content = ""
		} else if !util.IsInteger(line) { // Check if line is not chapter number
			content += "\n" + line // Append line into content of current chapter
		}
	}

	originalContentQueue.Enqueue(content) // send content each chapter to queue
	readFileCompleted = true              // Mark read file completely
}

/*
Dequeue content of each chapter and send to OpenAI api concurrently
Send n (limit) request to OpenAI in a time
Wait n (limit) request to OpenAI response and continue
*/
func summaryContent() {
	defer appWg.Done()

	var chapterIndex = 1
	for !readFileCompleted || !originalContentQueue.IsEmpty() {
		if originalContentQueue.Size() > OpenAiLimitRequest || readFileCompleted {
			var summaryContentArr = make([]string, OpenAiLimitRequest) // Array with size equal limit of amount of request send to OpenAI concurrently
			for i := range OpenAiLimitRequest {
				var originalContent = originalContentQueue.Dequeue() // Dequeue content of a chapter
				askWG.Add(1)
				go askAI(originalContent, chapterIndex, i, summaryContentArr) // Ask AI to summary content of chapter
				chapterIndex++
			}

			askWG.Wait() // Wait limit request to AI completely

			go writeFile(summaryContentArr) // Write n (limit) summary of chapter (answear from OpenAI) on file
		}
	}
}

// Send request to OpenAI to summary content of book's chapter
func askAI(content string, chaperIndex int, arrIndex int, summaryContentArr []string) {
	defer askWG.Done()
	openai.AskOpenAI(content, chaperIndex, arrIndex, summaryContentArr)
}

/*
Write summary of chapter append to file
*/
func writeFile(summaryContentArr []string) {
	f, err := os.OpenFile(SummaryFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
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

func check(e error) {
	if e != nil {
		panic(e)
	}
}
