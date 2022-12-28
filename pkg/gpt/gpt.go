package gpt

import (
	"context"
	"os"
	"regexp"
	"strings"

	gpt3 "github.com/PullRequestInc/go-gpt3"

	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/env"
	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/log"
)

var OpenAI gpt3.Client

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

	gptBlockedWord = "🏳️|🏳️‍🌈|🌈|lgbt|lesbian|gay|homosexual|homoseksual|bisexual|biseksual|transgender|fuck|sex"
	envBlockedWord := strings.TrimSpace(os.Getenv("CHATGPT_BLOCKED_WORD"))
	if len(envBlockedWord) > 0 {
		gptBlockedWord = gptBlockedWord + "|" + envBlockedWord
	}

	OpenAI = gpt3.NewClient(gptAPIKey)
}

func GPTResponse(question string) (response string, err error) {
	blockedWord := regexp.MustCompile(strings.ToLower(gptBlockedWord))

	if bool(blockedWord.MatchString(strings.ToLower(question))) {
		return "Cannot response to this question due to it contains blocked word 🥺", nil
	}

	gptResponse, err := OpenAI.CompletionWithEngine(
		context.Background(),
		gptModelName,
		gpt3.CompletionRequest{
			Prompt:           []string{question},
			MaxTokens:        gpt3.IntPtr(gptModelToken),
			Temperature:      gpt3.Float32Ptr(gptModelTemprature),
			TopP:             gpt3.Float32Ptr(gptModelTopP),
			PresencePenalty:  *gpt3.Float32Ptr(gptModelPenaltyPresence),
			FrequencyPenalty: *gpt3.Float32Ptr(gptModelPenaltyFreq),
		},
	)

	if err != nil {
		return "", err
	}

	buffResponse := strings.TrimSpace(gptResponse.Choices[0].Text)
	buffResponse = strings.TrimLeft(buffResponse, "?\n")
	buffResponse = strings.TrimLeft(buffResponse, "!\n")
	buffResponse = strings.TrimLeft(buffResponse, ":\n")
	buffResponse = strings.TrimLeft(buffResponse, "'\n")
	buffResponse = strings.TrimLeft(buffResponse, ".\n")
	buffResponse = strings.TrimLeft(buffResponse, "\n")

	if bool(blockedWord.MatchString(strings.ToLower(buffResponse))) {
		return "Cannot response to this question due to it contains blocked word 🥺", nil
	}

	return buffResponse, nil
}
