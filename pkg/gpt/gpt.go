package gpt

import (
	"context"
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
	OAIGPTModelName  string
	OAIGPTModelToken int
	OAIGPTModelTemperature,
	OAIGPTModelTopP,
	OAIGPTModelPenaltyPresence,
	OAIGPTModelPenaltyFreq float32
)

var (
	OGPTModelName string
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

	// -----------------------------------------------------------------------
	// OpenAI Configuration Environment
	// -----------------------------------------------------------------------
	OAIAPIKey, err := env.GetEnvString("OPENAI_API_KEY")
	if err != nil {
		if strings.ToLower(WAGPTEngine) == "openai" {
			log.Println(log.LogLevelFatal, "Error Parse Environment Variable for OpenAI API Key")
		}
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

	// -----------------------------------------------------------------------
	// Ollama Configuration Environment
	// -----------------------------------------------------------------------
	OGPTModelName, err = env.GetEnvString("OLLAMA_GPT_MODEL_NAME")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for Ollama GPT Model Name")
	}

	// -----------------------------------------------------------------------
	// GPT Engine Initialization
	// -----------------------------------------------------------------------
	switch strings.ToLower(WAGPTEngine) {
	case "openai":
		OAIClient = OpenAI.NewClient(OAIAPIKey)

	default:
		OClient, err = Ollama.ClientFromEnvironment()
		if err != nil {
			log.Println(log.LogLevelFatal, "Error, "+err.Error())
		}
	}
}

func GPT3Response(question string) (response string, err error) {
	if bool(WAGPTBlockedWordRegex.MatchString(question)) {
		return "Sorry, the AI can not response due to it is containing some blocked word 🥺", nil
	}

	switch strings.ToLower(WAGPTEngine) {
	case "openai":
		var OAIGPTResponseText string

		OAIGPTModel := regexp.MustCompile("\\b(?i)(" + "gpt-3\\.5" + ")")
		if bool(OAIGPTModel.MatchString(OAIGPTModelName)) {
			OAIGPTPrompt := OpenAI.ChatCompletionRequest{
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
		} else {
			OAIGPTPrompt := OpenAI.CompletionRequest{
				Model:            OAIGPTModelName,
				MaxTokens:        OAIGPTModelToken,
				Temperature:      OAIGPTModelTemperature,
				TopP:             OAIGPTModelTopP,
				PresencePenalty:  OAIGPTModelPenaltyPresence,
				FrequencyPenalty: OAIGPTModelPenaltyFreq,
				Prompt:           question,
			}

			OAIGPTResponse, err := OAIClient.CreateCompletion(
				context.Background(),
				OAIGPTPrompt,
			)

			if err != nil {
				return "", err
			}

			if len(OAIGPTResponse.Choices) > 0 {
				OAIGPTResponseText = OAIGPTResponse.Choices[0].Text
			}
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

		OGPTPrompt := &Ollama.ChatRequest{
			Model: OGPTModelName,
			Messages: []Ollama.Message{
				{
					Role:    "user",
					Content: question,
				},
			},
		}

		err := OClient.Chat(
			context.Background(),
			OGPTPrompt,
			func(OGPTResponse Ollama.ChatResponse) error {
				if len(OGPTResponse.Message.Content) > 0 {
					OGPTResponseText = OGPTResponse.Message.Content
				}
				return nil
			},
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
