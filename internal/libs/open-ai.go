package libs

import (
	"context"
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


// Include current Players from room
// Try to enforce the shape of the game state coming back from OpenAI
// and then use that defined shape to parse the game state into an object on the client

func getSystemMessage() openai.ChatCompletionMessage {
	// var playerStrings []string
	var compositeString string
	charName := "Grumby"
	// for _, pString := range playerStrings {
	// 	append(compositeString, pString + "\n")
	// }

	return openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleSystem,
		Content: "You are a Dungeon Master for a group of players.\n The players are:\n" +
			compositeString + "\n" +
			"You will respond to the players\" actions with a description of the results.\n" +
			"You will also update the value for GAME STATE.\n" +
			"The GAME STATE is JSON that consists of 2 values: Location (a string) and Enemies (a list of strings)\n\n" +
			"Here are some examples of how you might respond with a message and Game State for a player\"s action:\n" +
			charName + ": I search the corridor for enemies.\n\n" +
			"You search the corridor and notice a 2 goblins hiding behind a barrel holding a mace.\n" +
			"GAME STATE: { \"Location\": \"Dungeon Corridor\", \"Enemies\": [\"Goblin\", \"Goblin\"] }\n\n" +
			charName + ": I strike one of the goblins with my sword and succeed.\n\n" +
			"Your blade slices into the goblin\"s right arm and he drops his rusty mace.\n" +
			"GAME STATE: { \"Location\": \"Dungeon Corridor\", \"Enemies\": [\"Goblin\", \"wounded Goblin\"] }\n\n" +
			charName + ": Can I build a machine gun?\n\n" +
			"You lack the knowledge to build a machine gun and while you ponder it, the goblins escape.\n" +
			"GAME STATE: { \"Location\": \"Dungeon Corridor\", \"Enemies\": [] }\n\n" +
			charName + ": I go through the door at the end of the corridor.\n\n" +
			"You emerge into a large cavern with an idol in the center.\n" +
			"GAME STATE: { \"Location\": \"Large Cavern\", \"Enemies\": [] }\n\n",
	}
}

func GetChatCompletion(messages []openai.ChatCompletionMessage) (openai.ChatCompletionMessage, error) {
	client := getClient()
	systemMessage := getSystemMessage()
	msgsWithSystem := append([]openai.ChatCompletionMessage{systemMessage}, messages...)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "gpt-4",
			Messages: msgsWithSystem,
			Temperature: 1.2,
		},
	)

	if err != nil {
		return openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleSystem,
			Content: "Error: " + err.Error(),
		}, err
	}

	return resp.Choices[0].Message, nil
}