package gpt

import (
	"context"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"

	gpt "github.com/sashabaranov/go-gpt3"

	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/env"
	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/log"
)

var OpenAI *gpt.Client

var gptModelName string
var gptModelToken int

var (
	gptModelTemprature,
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

	gptModelTemprature, err = env.GetEnvFloat32("CHATGPT_MODEL_TEMPRATURE")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for ChatGPT Model Temprature")
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

	OpenAI = gpt.NewClient(gptAPIKey)
}

func GPTResponse(question string) (response string, err error) {
	blockedWord := regexp.MustCompile(strings.ToLower(gptBlockedWord))

	if bool(blockedWord.MatchString(strings.ToLower(question))) {
		return "Cannot response to this question due to it contains blocked word 🥺", nil
	}

	gptRequest := gpt.CompletionRequest{
		Model:            gptModelName,
		MaxTokens:        gptModelToken,
		Temperature:      gptModelTemprature,
		TopP:             gptModelTopP,
		PresencePenalty:  gptModelPenaltyPresence,
		FrequencyPenalty: gptModelPenaltyFreq,
		Prompt:           question,
	}

	gptResponse, err := OpenAI.CreateCompletionStream(
		context.Background(),
		gptRequest,
	)

	if err != nil {
		return "", err
	}
	defer gptResponse.Close()

	gptResponseWordPosition := 0
	gptResponseBuilder := strings.Builder{}

	for {
		gptResponseStream, err := gptResponse.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if len(gptResponseStream.Choices) > 0 {
			if gptResponseWordPosition == 0 {
				gptResponseBuilder.WriteString(strings.TrimLeft(gptResponseStream.Choices[0].Text, "\n"))
			} else {
				gptResponseBuilder.WriteString(gptResponseStream.Choices[0].Text)
			}

			gptResponseWordPosition++
		}
	}

	buffResponse := strings.TrimSpace(gptResponseBuilder.String())
	buffResponse = strings.TrimLeft(buffResponse, "?\n")
	buffResponse = strings.TrimLeft(buffResponse, "!\n")
	buffResponse = strings.TrimLeft(buffResponse, ":\n")
	buffResponse = strings.TrimLeft(buffResponse, "'\n")
	buffResponse = strings.TrimLeft(buffResponse, ".\n")
	buffResponse = strings.TrimLeft(buffResponse, "\n")

	return buffResponse, nil
}
