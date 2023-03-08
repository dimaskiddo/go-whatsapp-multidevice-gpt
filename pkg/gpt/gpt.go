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

var OAIGPTModelName string
var OAIGPTModelToken int

var (
	OAIGPTModelTemperature,
	OAIGPTModelTopP,
	OAIGPTModelPenaltyPresence,
	OAIGPTModelPenaltyFreq float32
)

var OAIGPTBlockedWord string

func init() {
	var err error

	OAIAPIKey, err := env.GetEnvString("OPENAI_API_KEY")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for OpenAI API Key")
	}

	OAIGPTModelName, err = env.GetEnvString("OPENAI_GPT_MODEL_NAME")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for OpenAI GPT Model Name")
	}

	OAIGPTModelToken, err = env.GetEnvInt("OPENAI_GPT_MODEL_TOKEN")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for OpenAI GPT Model Token")
	}

	OAIGPTModelTemperature, err = env.GetEnvFloat32("OPENAI_GPT_MODEL_TEMPERATURE")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for OpenAI GPT Model Temperature")
	}

	OAIGPTModelTopP, err = env.GetEnvFloat32("OPENAI_GPT_MODEL_TOP_P")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for OpenAI GPT Model Top P")
	}

	OAIGPTModelPenaltyPresence, err = env.GetEnvFloat32("OPENAI_GPT_MODEL_PENALTY_PRESENCE")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for OpenAI GPT Model Penalty Presence")
	}

	OAIGPTModelPenaltyFreq, err = env.GetEnvFloat32("OPENAI_GPT_MODEL_PENALTY_FREQUENCY")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for OpenAI GPT Model Penalty Frequency")
	}

	OAIGPTBlockedWord = "lgbt|lesbian|gay|homosexual|homoseksual|bisexual|biseksual|transgender|fuck|sex|masturbate|masturbasi|coli|colmek|jilmek|cock|penis|kontol|vagina|memek|porn|porno"
	envBlockedWord := strings.TrimSpace(os.Getenv("OPENAI_GPT_BLOCKED_WORD"))
	if len(envBlockedWord) > 0 {
		OAIGPTBlockedWord = OAIGPTBlockedWord + "|" + envBlockedWord
	}

	OAIClient = OpenAI.NewClient(OAIAPIKey)
}

func GPT3Response(question string) (response string, err error) {
	gptResponseBuilder := strings.Builder{}
	gptIsFirstWordFound := false

	gptBlockedWord := regexp.MustCompile(strings.ToLower(OAIGPTBlockedWord))
	if bool(gptBlockedWord.MatchString(strings.ToLower(question))) {
		return "Cannot response to this question due to it contains blocked word 🥺", nil
	}

	gptChatMode := regexp.MustCompile(strings.ToLower("gpt-3.5-turbo"))
	if bool(gptChatMode.MatchString(strings.ToLower(OAIGPTModelName))) {
		gptRequest := OpenAI.ChatCompletionRequest{
			Model:            OAIGPTModelName,
			MaxTokens:        OAIGPTModelToken,
			Temperature:      OAIGPTModelTemperature,
			TopP:             OAIGPTModelTopP,
			PresencePenalty:  OAIGPTModelPenaltyPresence,
			FrequencyPenalty: OAIGPTModelPenaltyFreq,
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

		for {
			gptResponseStream, err := gptResponse.Recv()
			if errors.Is(err, io.EOF) {
				break
			}

			if len(gptResponseStream.Choices) > 0 {
				gptWordResponse := gptResponseStream.Choices[0].Delta.Content
				gptWordResponse = strings.TrimLeft(gptWordResponse, "?\n")
				gptWordResponse = strings.TrimLeft(gptWordResponse, "!\n")
				gptWordResponse = strings.TrimLeft(gptWordResponse, ":\n")
				gptWordResponse = strings.TrimLeft(gptWordResponse, "'\n")
				gptWordResponse = strings.TrimLeft(gptWordResponse, ".\n")

				if !gptIsFirstWordFound && gptWordResponse != "\n" && len(strings.TrimSpace(gptWordResponse)) != 0 {
					gptIsFirstWordFound = true
				}

				if gptIsFirstWordFound {
					gptResponseBuilder.WriteString(gptResponseStream.Choices[0].Delta.Content)
				}
			}
		}
	} else {
		gptRequest := OpenAI.CompletionRequest{
			Model:            OAIGPTModelName,
			MaxTokens:        OAIGPTModelToken,
			Temperature:      OAIGPTModelTemperature,
			TopP:             OAIGPTModelTopP,
			PresencePenalty:  OAIGPTModelPenaltyPresence,
			FrequencyPenalty: OAIGPTModelPenaltyFreq,
			Prompt:           question,
		}

		gptResponse, err := OAIClient.CreateCompletionStream(
			context.Background(),
			gptRequest,
		)

		if err != nil {
			return "", err
		}
		defer gptResponse.Close()

		for {
			gptResponseStream, err := gptResponse.Recv()
			if errors.Is(err, io.EOF) {
				break
			}

			if len(gptResponseStream.Choices) > 0 {
				gptWordResponse := gptResponseStream.Choices[0].Text
				gptWordResponse = strings.TrimLeft(gptWordResponse, "?\n")
				gptWordResponse = strings.TrimLeft(gptWordResponse, "!\n")
				gptWordResponse = strings.TrimLeft(gptWordResponse, ":\n")
				gptWordResponse = strings.TrimLeft(gptWordResponse, "'\n")
				gptWordResponse = strings.TrimLeft(gptWordResponse, ".\n")

				if !gptIsFirstWordFound && gptWordResponse != "\n" && len(strings.TrimSpace(gptWordResponse)) != 0 {
					gptIsFirstWordFound = true
				}

				if gptIsFirstWordFound {
					gptResponseBuilder.WriteString(gptResponseStream.Choices[0].Text)
				}
			}
		}
	}

	if !gptIsFirstWordFound {
		return "Sorry, the AI can not response for this time. Please try again after a few moment. Thank you ! 🙈", nil
	}

	gptResponseBuffer := strings.TrimLeft(gptResponseBuilder.String(), "\n")
	gptResponseBuffer = strings.TrimSpace(gptResponseBuffer)

	return gptResponseBuffer, nil
}
