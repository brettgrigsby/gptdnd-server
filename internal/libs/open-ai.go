package libs

import (
	"context"
	"gptdnd-server/internal/models"
	"os"

	"github.com/joho/godotenv"

	openai "github.com/sashabaranov/go-openai"
)

func getClient() *openai.Client {
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")

	return openai.NewClient(apiKey)
}

func getSystemMessage() openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleSystem,
		Content: "You are a Dungeon Master for a group of players.",
	}
}

func GetChatCompletion(messages []openai.ChatCompletionMessage, players []models.Player) (string, error) {
	client := getClient()
	systemMessage := getSystemMessage()
	msgsWithSystem := append([]openai.ChatCompletionMessage{systemMessage}, messages...)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "gpt-4",
			Messages: msgsWithSystem,
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}