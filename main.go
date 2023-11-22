package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aataxe/chatgpt-adventures/openai"
)

func chat(ai *openai.Chat, prompt string, replyChan chan *openai.Message) {
	reply, err := ai.Conversation(prompt)
	if err != nil {
		fmt.Printf("Something went wrong...%v\n", err)
		time.Sleep(5 * time.Second)
		chat(ai, prompt, replyChan)
	} else {
		replyChan <- reply
	}
}

func main() {
	ai := openai.NewChat()
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("You:")

		prompt, _ := reader.ReadString('\n')
		prompt = strings.TrimSpace(prompt)

		replyChan := make(chan *openai.Message, 1)
		go chat(ai, prompt, replyChan)

		reply := <-replyChan

		fmt.Println("ChatGPT:")
		fmt.Println(reply.Content)
	}
}
