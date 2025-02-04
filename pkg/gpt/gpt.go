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

var (
	WAGPTEngine,
	WAGPTBlockedWord string
	WAGPTBlockedWordRegex *regexp.Regexp
)

var (
	OAIHost,
	OAIHostPath,
	OAIAPIKey string
)

var (
	GPTModelName,
	GPTModelPrompt string
	GPTModelToken int
	GPTModelTemperature,
	GPTModelTopP,
	GPTModelPenaltyPresence,
	GPTModelPenaltyFreq float32
)

const listBlockedWord string = "" +
	"lgbt|lesbian|gay|homosexual|homoseksual|bisexual|biseksual|transgender|" +
	"fuck|sex|ngentot|entot|ngewe|ewe|masturbate|masturbasi|coli|colmek|jilmek|" +
	"cock|penis|kontol|vagina|memek|porn|porno|bokep"

func init() {
	var err error

	// -----------------------------------------------------------------------
	// WhatsApp GPT Configuration Environment
	// -----------------------------------------------------------------------
	WAGPTBlockedWord = strings.TrimSpace(os.Getenv("WHATSAPP_GPT_BLOCKED_WORD"))
	if len(WAGPTBlockedWord) > 0 {
		WAGPTBlockedWordRegex = regexp.MustCompile("\\b(?i)(" + listBlockedWord + "|" + WAGPTBlockedWord + ")")
	} else {
		WAGPTBlockedWordRegex = regexp.MustCompile("\\b(?i)(" + listBlockedWord + ")")
	}

	// -----------------------------------------------------------------------
	// OpenAI Configuration Environment
	// -----------------------------------------------------------------------
	OAIHost, err = env.GetEnvString("OPENAI_HOST")
	if err != nil {
		OAIHost = "https://api.openai.com"
	}

	OAIHostPath, err = env.GetEnvString("OPENAI_HOST_PATH")
	if err != nil {
		OAIHostPath = "/v1"
	}

	OAIAPIKey, err = env.GetEnvString("OPENAI_API_KEY")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for OpenAI API Key")
	}

	// -----------------------------------------------------------------------
	// GPT Configuration Environment
	// -----------------------------------------------------------------------
	GPTModelName, err = env.GetEnvString("GPT_MODEL_NAME")
	if err != nil {
		GPTModelName = "gpt-3.5-turbo"
	}

	GPTModelPrompt, err = env.GetEnvString("GPT_MODEL_SYSTEM_PROMPT")
	if err != nil {
		GPTModelPrompt = ""
	}

	GPTModelToken, err = env.GetEnvInt("GPT_MODEL_TOKEN")
	if err != nil {
		GPTModelToken = 4096
	}

	GPTModelTemperature, err = env.GetEnvFloat32("GPT_MODEL_TEMPERATURE")
	if err != nil {
		GPTModelTemperature = 0
	}

	GPTModelTopP, err = env.GetEnvFloat32("GPT_MODEL_TOP_P")
	if err != nil {
		GPTModelTopP = 1
	}

	GPTModelPenaltyPresence, err = env.GetEnvFloat32("GPT_MODEL_PENALTY_PRESENCE")
	if err != nil {
		GPTModelPenaltyPresence = 0
	}

	GPTModelPenaltyFreq, err = env.GetEnvFloat32("GPT_MODEL_PENALTY_FREQUENCY")
	if err != nil {
		GPTModelPenaltyFreq = 0
	}

	// -----------------------------------------------------------------------
	// GPT Engine Initialization
	// -----------------------------------------------------------------------
	OAIConfig := OpenAI.DefaultConfig(OAIAPIKey)
	OAIConfig.BaseURL = OAIHost + OAIHostPath

	OAIClient = OpenAI.NewClientWithConfig(OAIConfig)
}

func GPTResponse(question string) (response string, err error) {
	if bool(WAGPTBlockedWordRegex.MatchString(question)) {
		return "Sorry, the AI can not response due to it is containing some blocked word 🥺", nil
	}

	isStream := new(bool)
	*isStream = true

	var OAIGPTResponseText string
	var OAIGPTChatCompletion []OpenAI.ChatCompletionMessage

	if len(strings.TrimSpace(GPTModelPrompt)) != 0 {
		OAIGPTChatCompletion = []OpenAI.ChatCompletionMessage{
			{
				Role:    OpenAI.ChatMessageRoleSystem,
				Content: GPTModelPrompt,
			},
			{
				Role:    OpenAI.ChatMessageRoleUser,
				Content: question,
			},
		}
	} else {
		OAIGPTChatCompletion = []OpenAI.ChatCompletionMessage{
			{
				Role:    OpenAI.ChatMessageRoleUser,
				Content: question,
			},
		}
	}

	OAIGPTPrompt := OpenAI.ChatCompletionRequest{
		Model:            GPTModelName,
		MaxTokens:        GPTModelToken,
		Temperature:      GPTModelTemperature,
		TopP:             GPTModelTopP,
		PresencePenalty:  GPTModelPenaltyPresence,
		FrequencyPenalty: GPTModelPenaltyFreq,
		Messages:         OAIGPTChatCompletion,
		Stream:           *isStream,
	}

	OAIGPTStream, err := OAIClient.CreateChatCompletionStream(
		context.Background(),
		OAIGPTPrompt,
	)

	if err != nil {
		return "", err
	}
	defer OAIGPTStream.Close()

	for {
		OAIGPTResponse, err := OAIGPTStream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return "", err
		}

		if len(OAIGPTResponse.Choices) > 0 {
			OAIGPTResponseText = OAIGPTResponseText + OAIGPTResponse.Choices[0].Delta.Content
		}
	}

	OAIGPTResponseBuffer := strings.TrimSpace(OAIGPTResponseText)
	OAIGPTResponseBuffer = strings.TrimLeft(OAIGPTResponseBuffer, "?\n")
	OAIGPTResponseBuffer = strings.TrimLeft(OAIGPTResponseBuffer, "!\n")
	OAIGPTResponseBuffer = strings.TrimLeft(OAIGPTResponseBuffer, ":\n")
	OAIGPTResponseBuffer = strings.TrimLeft(OAIGPTResponseBuffer, "'\n")
	OAIGPTResponseBuffer = strings.TrimLeft(OAIGPTResponseBuffer, ".\n")
	OAIGPTResponseBuffer = strings.TrimLeft(OAIGPTResponseBuffer, "\n")

	return OAIGPTResponseBuffer, nil
}
