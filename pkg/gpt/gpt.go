package gpt

import (
	"context"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"

	OpenAI "github.com/sashabaranov/go-openai"

	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/env"
	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/log"
)

var OAIClient *OpenAI.Client

var gptModelName string
var gptModelToken int

var (
	gptModelTemperature,
	gptModelTopP,
	gptModelPenaltyPresence,
	gptModelPenaltyFreq float32
)

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

	gptModelTemperature, err = env.GetEnvFloat32("CHATGPT_MODEL_TEMPERATURE")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for ChatGPT Model Temperature")
	}

	gptModelTopP, err = env.GetEnvFloat32("CHATGPT_MODEL_TOP_P")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for ChatGPT Model Top P")
	}

	gptModelPenaltyPresence, err = env.GetEnvFloat32("CHATGPT_MODEL_PENALTY_PRESENCE")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for ChatGPT Model Penalty Presence")
	}

	gptModelPenaltyFreq, err = env.GetEnvFloat32("CHATGPT_MODEL_PENALTY_FREQUENCY")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for ChatGPT Model Penalty Frequency")
	}

	gptBlockedWord = "lgbt|lesbian|gay|homosexual|homoseksual|bisexual|biseksual|transgender|fuck|sex|masturbate|masturbasi|coli|colmek|jilmek|cock|penis|kontol|vagina|memek|porn|porno"
	envBlockedWord := strings.TrimSpace(os.Getenv("CHATGPT_BLOCKED_WORD"))
	if len(envBlockedWord) > 0 {
		gptBlockedWord = gptBlockedWord + "|" + envBlockedWord
	}

	OAIClient = OpenAI.NewClient(gptAPIKey)
}

func GPT3Response(question string) (response string, err error) {
	blockedWord := regexp.MustCompile(strings.ToLower(gptBlockedWord))

	if bool(blockedWord.MatchString(strings.ToLower(question))) {
		return "Cannot response to this question due to it contains blocked word 🥺", nil
	}

	gptRequest := OpenAI.ChatCompletionRequest{
		Model:            gptModelName,
		MaxTokens:        gptModelToken,
		Temperature:      gptModelTemperature,
		TopP:             gptModelTopP,
		PresencePenalty:  gptModelPenaltyPresence,
		FrequencyPenalty: gptModelPenaltyFreq,
		Messages: []OpenAI.ChatCompletionMessage{
			{
				Role:    OpenAI.ChatMessageRoleUser,
				Content: question,
			},
		},
	}

	gptResponse, err := OAIClient.CreateChatCompletionStream(
		context.Background(),
		gptRequest,
	)

	if err != nil {
		return "", err
	}
	defer gptResponse.Close()

	gptResponseBuilder := strings.Builder{}
	gptIsFirstWordFound := false

	for {
		gptResponseStream, err := gptResponse.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if len(gptResponseStream.Choices) > 0 {
			gptWordResponse := gptResponseStream.Choices[0].Delta.Content
			if !gptIsFirstWordFound && gptWordResponse != "\n" && len(strings.TrimSpace(gptWordResponse)) != 0 {
				gptIsFirstWordFound = true
			}

			if gptIsFirstWordFound {
				gptResponseBuilder.WriteString(gptResponseStream.Choices[0].Delta.Content)
			}
		}
	}

	if !gptIsFirstWordFound {
		return "Sorry, can't response this question for this time. Please try again after a few moment. Thank you !", nil
	}

	gptResponseCleaner := strings.TrimSpace(gptResponseBuilder.String())
	gptResponseCleaner = strings.TrimLeft(gptResponseCleaner, "?\n")
	gptResponseCleaner = strings.TrimLeft(gptResponseCleaner, "!\n")
	gptResponseCleaner = strings.TrimLeft(gptResponseCleaner, ":\n")
	gptResponseCleaner = strings.TrimLeft(gptResponseCleaner, "'\n")
	gptResponseCleaner = strings.TrimLeft(gptResponseCleaner, ".\n")
	gptResponseCleaner = strings.TrimLeft(gptResponseCleaner, "\n")

	return gptResponseCleaner, nil
}
