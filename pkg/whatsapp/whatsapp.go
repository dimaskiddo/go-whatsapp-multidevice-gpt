package whatsapp

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"
	waproto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/env"
	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/gpt"
	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/log"
)

var WhatsAppDatastore *sqlstore.Container
var WhatsAppClient *whatsmeow.Client

var chatGPTTag string

func init() {
	var err error

	dbType, err := env.GetEnvString("WHATSAPP_DATASTORE_TYPE")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for WhatsApp Client Datastore Type")
	}

	dbURI, err := env.GetEnvString("WHATSAPP_DATASTORE_URI")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for WhatsApp Client Datastore URI")
	}

	datastore, err := sqlstore.New(dbType, dbURI, nil)
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Connect WhatsApp Client Datastore")
	}

	chatGPTTag, err = env.GetEnvString("WHATSAPP_CHATGPT_TAG")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for WhatsApp Client ChatGPT Tag")
	}

	WhatsAppDatastore = datastore
}

func WhatsAppInitClient(device *store.Device) {
	var err error

	if WhatsAppClient == nil {
		if device == nil {
			// Initialize New WhatsApp Client Device in Datastore
			device = WhatsAppDatastore.NewDevice()
		}

		// Set Client Properties
		store.DeviceProps.Os = proto.String("Go WhatsApp Multi-Device GPT")
		store.DeviceProps.PlatformType = waproto.DeviceProps_DESKTOP.Enum()
		store.DeviceProps.RequireFullSync = proto.Bool(false)

		// Set Client Versions
		version.Major, err = env.GetEnvInt("WHATSAPP_VERSION_MAJOR")
		if err == nil {
			store.DeviceProps.Version.Primary = proto.Uint32(uint32(version.Major))
		}
		version.Minor, err = env.GetEnvInt("WHATSAPP_VERSION_MINOR")
		if err == nil {
			store.DeviceProps.Version.Secondary = proto.Uint32(uint32(version.Minor))
		}
		version.Patch, err = env.GetEnvInt("WHATSAPP_VERSION_PATCH")
		if err == nil {
			store.DeviceProps.Version.Tertiary = proto.Uint32(uint32(version.Patch))
		}

		// Initialize New WhatsApp Client
		WhatsAppClient = whatsmeow.NewClient(device, nil)

		// Set WhatsApp Client Auto Reconnect
		WhatsAppClient.EnableAutoReconnect = true

		// Set WhatsApp Client Auto Trust Identity
		WhatsAppClient.AutoTrustIdentity = true
	}
}

func WhatsAppGenerateQR(qrChan <-chan whatsmeow.QRChannelItem) (string, int) {
	qrChanCode := make(chan string)
	qrChanTimeout := make(chan int)

	// Get QR Code Data and Timeout
	go func() {
		for evt := range qrChan {
			if evt.Event == "code" {
				qrChanCode <- evt.Code
				qrChanTimeout <- int(evt.Timeout.Seconds())
			}
		}
	}()

	// Return QR Code Data and Timeout Information
	return <-qrChanCode, <-qrChanTimeout
}

func WhatsAppLogin() (string, int, error) {
	if WhatsAppClient != nil {
		// Make Sure WebSocket Connection is Disconnected
		WhatsAppClient.Disconnect()

		if WhatsAppClient.Store.ID == nil {
			// Device ID is not Exist
			// Generate QR Code
			qrChanGenerate, _ := WhatsAppClient.GetQRChannel(context.Background())

			// Connect WebSocket while Initialize QR Code Data to be Sent
			err := WhatsAppClient.Connect()
			if err != nil {
				return "", 0, err
			}

			// Set WhatsApp Client Presence to Available
			_ = WhatsAppClient.SendPresence(types.PresenceAvailable)

			// Get Generated QR Code and Timeout Information
			qrString, qrTimeout := WhatsAppGenerateQR(qrChanGenerate)

			// Print QR Code in Terminal
			return qrString, qrTimeout, nil
		} else {
			// Device ID is Exist
			// Reconnect WebSocket
			err := WhatsAppReconnect()
			if err != nil {
				return "", 0, err
			}

			return "WhatsApp Client is Reconnected", 0, nil
		}
	}

	// Return Error WhatsApp Client is not Valid
	return "", 0, errors.New("WhatsApp Client is not Valid")
}

func WhatsAppReconnect() error {
	if WhatsAppClient != nil {
		// Make Sure WebSocket Connection is Disconnected
		WhatsAppClient.Disconnect()

		// Make Sure Store ID is not Empty
		// To do Reconnection
		if WhatsAppClient != nil {
			err := WhatsAppClient.Connect()
			if err != nil {
				return err
			}

			// Set WhatsApp Client Presence to Available
			_ = WhatsAppClient.SendPresence(types.PresenceAvailable)

			return nil
		}

		return errors.New("WhatsApp Client Store ID is Empty, Please Re-Login and Scan QR Code Again")
	}

	return errors.New("WhatsApp Client is not Valid")
}

func WhatsAppLogout() error {
	if WhatsAppClient != nil {
		// Make Sure Store ID is not Empty
		if WhatsAppClient.Store.ID != nil {
			var err error

			// Set WhatsApp Client Presence to Unavailable
			_ = WhatsAppClient.SendPresence(types.PresenceUnavailable)

			// Logout WhatsApp Client and Disconnect from WebSocket
			err = WhatsAppClient.Logout()
			if err != nil {
				// Force Disconnect
				WhatsAppClient.Disconnect()

				// Manually Delete Device from Datastore Store
				err = WhatsAppClient.Store.Delete()
				if err != nil {
					return err
				}
			}

			// Free WhatsApp Client Map
			WhatsAppClient = nil
			return nil
		}

		return errors.New("WhatsApp Client Store ID is Empty, Please Re-Login and Scan QR Code Again")
	}

	// Return Error WhatsApp Client is not Valid
	return errors.New("WhatsApp Client is not Valid")
}

func WhatsAppComposeStatus(rjid types.JID, isComposing bool, isAudio bool) {
	// Set Compose Status
	var typeCompose types.ChatPresence
	if isComposing {
		typeCompose = types.ChatPresenceComposing
	} else {
		typeCompose = types.ChatPresencePaused
	}

	// Set Compose Media Audio (Recording) or Text (Typing)
	var typeComposeMedia types.ChatPresenceMedia
	if isAudio {
		typeComposeMedia = types.ChatPresenceMediaAudio
	} else {
		typeComposeMedia = types.ChatPresenceMediaText
	}

	// Send Chat Compose Status
	_ = WhatsAppClient.SendChatPresence(rjid, typeCompose, typeComposeMedia)
}

func WhatsAppSendGPTResponse(ctx context.Context, event *events.Message, response string) (string, error) {
	if WhatsAppClient != nil {
		var err error

		// Make Sure WhatsApp Client is OK
		if WhatsAppClient.IsConnected() && WhatsAppClient.IsLoggedIn() {
			rJID := event.Info.Chat

			// Set Chat Presence
			WhatsAppComposeStatus(rJID, true, false)
			defer WhatsAppComposeStatus(rJID, false, false)

			// Compose WhatsApp Proto
			var msgContent *waproto.Message
			msgId := whatsmeow.GenerateMessageID()

			if strings.ContainsRune(rJID.String(), '-') {
				msgContent = &waproto.Message{
					Conversation: proto.String(response),
				}
			} else {
				msgContent = &waproto.Message{
					ExtendedTextMessage: &waproto.ExtendedTextMessage{
						Text: proto.String(response),
						ContextInfo: &waproto.ContextInfo{
							StanzaId:      proto.String(event.Info.ID),
							Participant:   proto.String(event.Info.Sender.String()),
							QuotedMessage: event.Message,
						},
					},
				}
			}

			// Send WhatsApp Message Proto
			_, err = WhatsAppClient.SendMessage(ctx, rJID, msgId, msgContent)
			if err != nil {
				return "", err
			}

			return msgId, nil
		} else {
			return "", errors.New("WhatsApp Client is not Connected or Logged-in")
		}
	}

	// Return Error WhatsApp Client is not Valid
	return "", errors.New("WhatsApp Client is not Valid")
}

func WhatsAppHandler(event interface{}) {
	switch evt := event.(type) {
	case *events.Message:
		if evt.Info.MediaType == "" {
			realRJID := evt.Info.Chat.String()

			var maskRJID string
			if strings.ContainsRune(realRJID, '-') {
				splitRJID := strings.Split(realRJID, "-")

				realRJID = splitRJID[0]
				maskRJID = realRJID[0:len(realRJID)-4] + "xxxx" + "-" + splitRJID[1]
			} else {
				splitRJID := strings.Split(realRJID, "@")

				realRJID = splitRJID[0]
				maskRJID = realRJID[0:len(realRJID)-4] + "xxxx" + "@" + splitRJID[1]
			}

			rMessage := strings.TrimSpace(evt.Message.GetConversation())

			if strings.Contains(rMessage, chatGPTTag+" ") {
				splitByTag := strings.Split(rMessage, chatGPTTag+" ")
				question := strings.TrimSpace(splitByTag[1])

				if len(strings.TrimSpace(question)) > 0 {
					log.Println(log.LogLevelInfo, "-== Incomming Question ==-")
					log.Println(log.LogLevelInfo, "From     : "+maskRJID)
					log.Println(log.LogLevelInfo, "Question : "+question)

					response, err := gpt.GPT3Response(question)
					if err != nil {
						response = "Failed to Get Response, Got Timeout from OpenAI GPT 🙈"
					}

					if len(strings.TrimSpace(response)) > 0 {
						_, err = WhatsAppSendGPTResponse(context.Background(), evt, response)
						if err != nil {
							log.Println(log.LogLevelError, "Failed to Send OpenAI GPT Response")
						}
					}
				}
			}
		}
	}
}
