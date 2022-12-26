package chatgpt

import (
	"context"
	"strings"

	gpt "github.com/PullRequestInc/go-gpt3"

	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/env"
	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/log"
)

var gptAPIKey string

func init() {
	var err error

	gptAPIKey, err = env.GetEnvString("CHATGPT_API_KEY")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for ChatGPT API Key")
	}
}

func ChatGPTResponse(question string) (response string, err error) {
	openai := gpt.NewClient(gptAPIKey)

	responseBuilder := strings.Builder{}

	err = openai.CompletionStreamWithEngine(
		context.Background(),
		gpt.TextDavinci003Engine,
		gpt.CompletionRequest{
			Prompt:      []string{question},
			MaxTokens:   gpt.IntPtr(3000),
			Temperature: gpt.Float32Ptr(0),
		},
		func(gptResponse *gpt.CompletionResponse) {
			responseBuilder.WriteString(gptResponse.Choices[0].Text)
		},
	)

	if err != nil {
		return "", err
	}

	response = strings.TrimLeft(responseBuilder.String(), "\n")
	return response, nil
}
