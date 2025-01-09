package gpt

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	Ollama "github.com/ollama/ollama/api"
	OpenAI "github.com/sashabaranov/go-openai"

	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/env"
	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/log"
)

var OAIClient *OpenAI.Client
var OClient *Ollama.Client

var (
	WAGPTEngine,
	WAGPTBlockedWord string
	WAGPTBlockedWordRegex *regexp.Regexp
)

var (
	OAIHost,
	OAIHostPath,
	OAIAPIKey,
	OAIGPTModelName,
	OAIGPTModelPrompt string
	OAIGPTModelToken int
	OAIGPTModelTemperature,
	OAIGPTModelTopP,
	OAIGPTModelPenaltyPresence,
	OAIGPTModelPenaltyFreq float32
)

var (
	OHost,
	OHostPath,
	OGPTModelName,
	OGPTModelPrompt string
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
	WAGPTEngine, err = env.GetEnvString("WHATSAPP_GPT_ENGINE")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for WhatsApp GPT Engine")
	}

	WAGPTBlockedWord = strings.TrimSpace(os.Getenv("WHATSAPP_GPT_BLOCKED_WORD"))
	if len(WAGPTBlockedWord) > 0 {
		WAGPTBlockedWordRegex = regexp.MustCompile("\\b(?i)(" + listBlockedWord + "|" + WAGPTBlockedWord + ")")
	} else {
		WAGPTBlockedWordRegex = regexp.MustCompile("\\b(?i)(" + listBlockedWord + ")")
	}

	switch strings.ToLower(WAGPTEngine) {
	case "openai":
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

		OAIGPTModelName, err = env.GetEnvString("OPENAI_GPT_MODEL_NAME")
		if err != nil {
			OAIGPTModelName = "gpt-3.5-turbo"
		}

		OAIGPTModelPrompt, err = env.GetEnvString("OPENAI_GPT_MODEL_SYSTEM_PROMPT")
		if err != nil {
			OAIGPTModelPrompt = ""
		}

		OAIGPTModelToken, err = env.GetEnvInt("OPENAI_GPT_MODEL_TOKEN")
		if err != nil {
			OAIGPTModelToken = 4096
		}

		OAIGPTModelTemperature, err = env.GetEnvFloat32("OPENAI_GPT_MODEL_TEMPERATURE")
		if err != nil {
			OAIGPTModelTemperature = 0
		}

		OAIGPTModelTopP, err = env.GetEnvFloat32("OPENAI_GPT_MODEL_TOP_P")
		if err != nil {
			OAIGPTModelTopP = 1
		}

		OAIGPTModelPenaltyPresence, err = env.GetEnvFloat32("OPENAI_GPT_MODEL_PENALTY_PRESENCE")
		if err != nil {
			OAIGPTModelPenaltyPresence = 0
		}

		OAIGPTModelPenaltyFreq, err = env.GetEnvFloat32("OPENAI_GPT_MODEL_PENALTY_FREQUENCY")
		if err != nil {
			OAIGPTModelPenaltyFreq = 0
		}

	default:
		// -----------------------------------------------------------------------
		// Ollama Configuration Environment
		// -----------------------------------------------------------------------
		OHost, err = env.GetEnvString("OLLAMA_HOST")
		if err != nil {
			log.Println(log.LogLevelFatal, "Error Parse Environment Variable for Ollama Host")
		}

		OHostPath, err = env.GetEnvString("OLLAMA_HOST_PATH")
		if err != nil {
			OHostPath = "/"
		}

		OGPTModelName, err = env.GetEnvString("OLLAMA_GPT_MODEL_NAME")
		if err != nil {
			log.Println(log.LogLevelFatal, "Error Parse Environment Variable for Ollama GPT Model Name")
		}

		OGPTModelPrompt, err = env.GetEnvString("OLLAMA_GPT_MODEL_SYSTEM_PROMPT")
		if err != nil {
			OGPTModelPrompt = ""
		}
	}

	// -----------------------------------------------------------------------
	// GPT Engine Initialization
	// -----------------------------------------------------------------------
	switch strings.ToLower(WAGPTEngine) {
	case "openai":
		OAIConfig := OpenAI.DefaultConfig(OAIAPIKey)
		OAIConfig.BaseURL = OAIHost + OAIHostPath

		OAIClient = OpenAI.NewClientWithConfig(OAIConfig)

	default:
		var OHostPort string

		OHostSchema, OHostURL, isOK := strings.Cut(OHost, "://")

		if !isOK {
			OHostSchema = "http"
			OHostURL = OHost
			OHostPort = "11434"
		}

		switch OHostSchema {
		case "http":
			OHostPort = "80"

		case "https":
			OHostPort = "443"
		}

		OClient = Ollama.NewClient(&url.URL{
			Scheme: OHostSchema,
			Host:   net.JoinHostPort(OHostURL, OHostPort),
			Path:   OHostPath,
		}, http.DefaultClient)
	}
}

func GPT3Response(question string) (response string, err error) {
	if bool(WAGPTBlockedWordRegex.MatchString(question)) {
		return "Sorry, the AI can not response due to it is containing some blocked word 🥺", nil
	}

	switch strings.ToLower(WAGPTEngine) {
	case "openai":
		var OAIGPTResponseText string
		var OAIGPTChatCompletion []OpenAI.ChatCompletionMessage

		if len(strings.TrimSpace(OAIGPTModelPrompt)) != 0 {
			OAIGPTChatCompletion = []OpenAI.ChatCompletionMessage{
				{
					Role:    OpenAI.ChatMessageRoleSystem,
					Content: OAIGPTModelPrompt,
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
			Model:            OAIGPTModelName,
			MaxTokens:        OAIGPTModelToken,
			Temperature:      OAIGPTModelTemperature,
			TopP:             OAIGPTModelTopP,
			PresencePenalty:  OAIGPTModelPenaltyPresence,
			FrequencyPenalty: OAIGPTModelPenaltyFreq,
			Messages:         OAIGPTChatCompletion,
			Stream:           false,
		}

		OAIGPTResponse, err := OAIClient.CreateChatCompletion(
			context.Background(),
			OAIGPTPrompt,
		)

		if err != nil {
			return "", err
		}

		if len(OAIGPTResponse.Choices) > 0 {
			OAIGPTResponseText = OAIGPTResponse.Choices[0].Message.Content
		}

		OAIGPTResponseBuffer := strings.TrimSpace(OAIGPTResponseText)
		OAIGPTResponseBuffer = strings.TrimLeft(OAIGPTResponseBuffer, "?\n")
		OAIGPTResponseBuffer = strings.TrimLeft(OAIGPTResponseBuffer, "!\n")
		OAIGPTResponseBuffer = strings.TrimLeft(OAIGPTResponseBuffer, ":\n")
		OAIGPTResponseBuffer = strings.TrimLeft(OAIGPTResponseBuffer, "'\n")
		OAIGPTResponseBuffer = strings.TrimLeft(OAIGPTResponseBuffer, ".\n")
		OAIGPTResponseBuffer = strings.TrimLeft(OAIGPTResponseBuffer, "\n")

		if bool(WAGPTBlockedWordRegex.MatchString(OAIGPTResponseBuffer)) {
			return "Sorry, the AI can not response due to it is containing some blocked word 🥺", nil
		}

		return OAIGPTResponseBuffer, nil

	default:
		var OGPTResponseText string
		var OGPTChatCompletion []Ollama.Message

		isStream := new(bool)
		*isStream = false

		if len(strings.TrimSpace(OGPTModelPrompt)) != 0 {
			OGPTChatCompletion = []Ollama.Message{
				{
					Role:    "system",
					Content: OGPTModelPrompt,
				},
				{
					Role:    "user",
					Content: question,
				},
			}
		} else {
			OGPTChatCompletion = []Ollama.Message{
				{
					Role:    "user",
					Content: question,
				},
			}
		}

		OGPTPrompt := &Ollama.ChatRequest{
			Model:    OGPTModelName,
			Messages: OGPTChatCompletion,
			Stream:   isStream,
		}

		OGTPResponseFunc := func(OGPTResponse Ollama.ChatResponse) error {
			OGPTResponseText = OGPTResponse.Message.Content
			return nil
		}

		err := OClient.Chat(
			context.Background(),
			OGPTPrompt,
			OGTPResponseFunc,
		)

		if err != nil {
			return "", err
		}

		OGPTResponseBuffer := strings.TrimSpace(OGPTResponseText)
		OGPTResponseBuffer = strings.TrimLeft(OGPTResponseBuffer, "?\n")
		OGPTResponseBuffer = strings.TrimLeft(OGPTResponseBuffer, "!\n")
		OGPTResponseBuffer = strings.TrimLeft(OGPTResponseBuffer, ":\n")
		OGPTResponseBuffer = strings.TrimLeft(OGPTResponseBuffer, "'\n")
		OGPTResponseBuffer = strings.TrimLeft(OGPTResponseBuffer, ".\n")
		OGPTResponseBuffer = strings.TrimLeft(OGPTResponseBuffer, "\n")

		if bool(WAGPTBlockedWordRegex.MatchString(OGPTResponseBuffer)) {
			return "Sorry, the AI can not response due to it is containing some blocked word 🥺", nil
		}

		return OGPTResponseBuffer, nil
	}
}
