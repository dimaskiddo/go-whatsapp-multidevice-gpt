package chatgpt

import (
	"context"
	"regexp"
	"strings"

	gpt "github.com/PullRequestInc/go-gpt3"

	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/env"
	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/log"
)

var OpenAI gpt.Client

var gptModelName string
var gptModelToken int
var gptModelTemprature float32

var gptBlockedWord string

func init() {
	var err error

	gptAPIKey, err := env.GetEnvString("CHATGPT_API_KEY")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for ChatGPT API Key")
	}

	gptModelName, err = env.GetEnvString("CHATGPT_MODEL_NAME")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for ChatGPT Model Name")
	}

	gptModelToken, err = env.GetEnvInt("CHATGPT_MODEL_TOKEN")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for ChatGPT Model Token")
	}

	gptModelTemprature, err = env.GetEnvFloat32("CHATGPT_MODEL_TEMPRATURE")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for ChatGPT Model Temprature")
	}

	gptBlockedWord, err = env.GetEnvString("CHATGPT_BLOCKED_WORD")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for ChatGPT Model Name")
	}

	OpenAI = gpt.NewClient(gptAPIKey)
}

func ChatGPTResponse(question string) (response string, err error) {
	blockedWord := regexp.MustCompile(strings.ToLower(gptBlockedWord))
	if bool(blockedWord.MatchString(strings.ToLower(question))) {
		return "Cannot response to this question due to it contains blocked word 🤬", nil
	}

	gptResponse, err := OpenAI.CompletionWithEngine(
		context.Background(),
		gptModelName,
		gpt.CompletionRequest{
			Prompt:      []string{question},
			MaxTokens:   gpt.IntPtr(gptModelToken),
			Temperature: gpt.Float32Ptr(gptModelTemprature),
		},
	)

	if err != nil {
		return "", err
	}

	buffResponse := strings.TrimSpace(gptResponse.Choices[0].Text)
	buffResponse = strings.TrimLeft(buffResponse, "?\n")
	buffResponse = strings.TrimLeft(buffResponse, "!\n")
	buffResponse = strings.TrimLeft(buffResponse, "'\n")
	buffResponse = strings.TrimLeft(buffResponse, ".\n")
	buffResponse = strings.TrimLeft(buffResponse, "\n")

	return buffResponse, nil
}
