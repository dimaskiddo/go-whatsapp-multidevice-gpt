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

func WhatsAppIsClientOK() error {
	// Make Sure WhatsApp Client is Connected
	if !WhatsAppClient.IsConnected() {
		return errors.New("WhatsApp Client is not Connected")
	}

	// Make Sure WhatsApp Client is Logged In
	if !WhatsAppClient.IsLoggedIn() {
		return errors.New("WhatsApp Client is not Logged In")
	}

	return nil
}

func WhatsAppGetJID(id string) types.JID {
	if WhatsAppClient != nil {
		var ids []string

		ids = append(ids, "+"+id)
		infos, err := WhatsAppClient.IsOnWhatsApp(ids)
		if err == nil {
			// If WhatsApp ID is Registered Then
			// Return ID Information
			if infos[0].IsIn {
				return infos[0].JID
			}
		}
	}

	// Return Empty ID Information
	return types.EmptyJID
}

func WhatsAppComposeJID(id string) types.JID {
	// Decompose WhatsApp ID First Before Recomposing
	id = WhatsAppDecomposeJID(id)

	// Check if ID is Group or Not By Detecting '-' for Old Group ID
	// Or By ID Length That Should be 18 Digits or More
	if strings.ContainsRune(id, '-') || len(id) >= 18 {
		// Return New Group User JID
		return types.NewJID(id, types.GroupServer)
	}

	// Return New Standard User JID
	return types.NewJID(id, types.DefaultUserServer)
}

func WhatsAppDecomposeJID(id string) string {
	// Check if WhatsApp ID Contains '@' Symbol
	if strings.ContainsRune(id, '@') {
		// Split WhatsApp ID Based on '@' Symbol
		// and Get Only The First Section Before The Symbol
		buffers := strings.Split(id, "@")
		id = buffers[0]
	}

	// Check if WhatsApp ID First Character is '+' Symbol
	if id[0] == '+' {
		// Remove '+' Symbol from WhatsApp ID
		id = id[1:]
	}

	return id
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
		err = WhatsAppIsClientOK()
		if err != nil {
			return "", err
		}

		// Compose New Remote JID
		remoteJID := WhatsAppComposeJID(event.Info.Sender.User)
		if WhatsAppGetJID(remoteJID.String()).IsEmpty() {
			return "", errors.New("WhatsApp Personal ID is Not Registered")
		}

		// Set Chat Presence
		defer WhatsAppComposeStatus(remoteJID, false, false)

		// Compose WhatsApp Proto
		msgId := whatsmeow.GenerateMessageID()
		msgContent := &waproto.Message{
			ExtendedTextMessage: &waproto.ExtendedTextMessage{
				Text: proto.String(response),
				ContextInfo: &waproto.ContextInfo{
					StanzaId:      proto.String(event.Info.ID),
					Participant:   proto.String(remoteJID.String()),
					QuotedMessage: event.Message,
				},
			},
		}

		// Send WhatsApp Message Proto
		_, err = WhatsAppClient.SendMessage(ctx, remoteJID, msgId, msgContent)
		if err != nil {
			return "", err
		}

		return msgId, nil
	}

	// Return Error WhatsApp Client is not Valid
	return "", errors.New("WhatsApp Client is not Valid")
}

func WhatsAppHandler(event interface{}) {
	switch evt := event.(type) {
	case *events.Message:
		var err error
		var response string

		if evt.Info.MediaType == "" {
			realRJID := evt.Info.Sender.User
			maskRJID := realRJID[0:len(realRJID)-4] + "xxxx"

			if realRJID != WhatsAppClient.Store.ID.User {
				rMessage := strings.TrimSpace(evt.Message.GetConversation())

				if strings.Contains(rMessage, chatGPTTag+" ") {
					splitByTag := strings.Split(rMessage, chatGPTTag+" ")
					question := strings.TrimSpace(splitByTag[1])

					log.Println(log.LogLevelInfo, "-== Incomming Question ==-")
					log.Println(log.LogLevelInfo, "From     : "+maskRJID)
					log.Println(log.LogLevelInfo, "Question : "+question)

					response, err = gpt.GPTResponse(question)
					if err != nil {
						response = "Failed to Get Reponse, Got Timeout from OpenAI GPT 🙈"
					}

					_, err = WhatsAppSendGPTResponse(context.Background(), evt, response)
					if err != nil {
						log.Println(log.LogLevelError, "Failed to Send OpenAI GPT Response")
					}
				}
			}
		}
	}
}
