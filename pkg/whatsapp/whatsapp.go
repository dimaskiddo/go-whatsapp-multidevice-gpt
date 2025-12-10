package whatsapp

import (
	"context"
	"errors"
	"regexp"
	"runtime"
	"strings"

	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"
	wabin "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/proto/waE2E"
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

var (
	WhatsAppClientProxyURL,
	WhatsAppGPTTag string
)

var WhatsAppGPTTagRegex *regexp.Regexp

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

	datastore, err := sqlstore.New(context.Background(), dbType, dbURI, nil)
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Connect WhatsApp Client Datastore")
	}

	WhatsAppClientProxyURL, _ = env.GetEnvString("WHATSAPP_CLIENT_PROXY_URL")

	WhatsAppGPTTag, err = env.GetEnvString("WHATSAPP_GPT_TAG")
	if err != nil {
		log.Println(log.LogLevelFatal, "Error Parse Environment Variable for WhatsApp GPT Tag")
	}

	WhatsAppGPTTag = strings.TrimSpace(strings.ToLower(WhatsAppGPTTag))
	WhatsAppGPTTagRegex = regexp.MustCompile("\\b(?i)(" + WhatsAppGPTTag + " " + ")")

	WhatsAppDatastore = datastore
}

func WhatsAppInitClient(device *store.Device) {
	var err error
	wabin.IndentXML = true

	if WhatsAppClient == nil {
		if device == nil {
			// Initialize New WhatsApp Client Device in Datastore
			device = WhatsAppDatastore.NewDevice()
		}

		// Set Client Properties
		store.DeviceProps.Os = proto.String(WhatsAppGetUserOS())
		store.DeviceProps.PlatformType = WhatsAppGetUserAgent("chrome").Enum()
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

		// Set WhatsApp Client Proxy Address if Proxy URL is Provided
		if len(WhatsAppClientProxyURL) > 0 {
			WhatsAppClient.SetProxyAddress(WhatsAppClientProxyURL)
		}

		// Set WhatsApp Client Auto Reconnect
		WhatsAppClient.EnableAutoReconnect = true

		// Set WhatsApp Client Auto Trust Identity
		WhatsAppClient.AutoTrustIdentity = true
	}
}

func WhatsAppGetUserAgent(agentType string) waCompanionReg.DeviceProps_PlatformType {
	switch strings.ToLower(agentType) {
	case "desktop":
		return waCompanionReg.DeviceProps_DESKTOP
	case "mac":
		return waCompanionReg.DeviceProps_CATALINA
	case "android":
		return waCompanionReg.DeviceProps_ANDROID_AMBIGUOUS
	case "android-phone":
		return waCompanionReg.DeviceProps_ANDROID_PHONE
	case "andorid-tablet":
		return waCompanionReg.DeviceProps_ANDROID_TABLET
	case "ios-phone":
		return waCompanionReg.DeviceProps_IOS_PHONE
	case "ios-catalyst":
		return waCompanionReg.DeviceProps_IOS_CATALYST
	case "ipad":
		return waCompanionReg.DeviceProps_IPAD
	case "wearos":
		return waCompanionReg.DeviceProps_WEAR_OS
	case "ie":
		return waCompanionReg.DeviceProps_IE
	case "edge":
		return waCompanionReg.DeviceProps_EDGE
	case "chrome":
		return waCompanionReg.DeviceProps_CHROME
	case "safari":
		return waCompanionReg.DeviceProps_SAFARI
	case "firefox":
		return waCompanionReg.DeviceProps_FIREFOX
	case "opera":
		return waCompanionReg.DeviceProps_OPERA
	case "uwp":
		return waCompanionReg.DeviceProps_UWP
	case "aloha":
		return waCompanionReg.DeviceProps_ALOHA
	case "tv-tcl":
		return waCompanionReg.DeviceProps_TCL_TV
	default:
		return waCompanionReg.DeviceProps_UNKNOWN
	}
}

func WhatsAppGetUserOS() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows"
	case "darwin":
		return "macOS"
	default:
		return "Linux"
	}
}

func WhatsAppLogin(jid string) (string, int, error) {
	if WhatsAppClient != nil {
		// Make Sure WebSocket Connection is Disconnected
		WhatsAppClient.Disconnect()

		if WhatsAppClient.Store.ID == nil {
			// Connect WebSocket while also Requesting Pairing Code
			err := WhatsAppClient.Connect()
			if err != nil {
				return "", 0, err
			}

			// Request Pairing Code
			code, err := WhatsAppClient.PairPhone(context.Background(), jid, true, whatsmeow.PairClientChrome, "Chrome ("+WhatsAppGetUserOS()+")")
			if err != nil {
				return "", 0, err
			}

			return code, 160, nil
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
			WhatsAppPresence(false)

			// Logout WhatsApp Client and Disconnect from WebSocket
			err = WhatsAppClient.Logout(context.Background())
			if err != nil {
				// Force Disconnect
				WhatsAppClient.Disconnect()

				// Manually Delete Device from Datastore Store
				err = WhatsAppClient.Store.Delete(context.Background())
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

func WhatsAppPresence(isAvailable bool) {
	if isAvailable {
		_ = WhatsAppClient.SendPresence(context.Background(), types.PresenceAvailable)
	} else {
		_ = WhatsAppClient.SendPresence(context.Background(), types.PresenceUnavailable)
	}
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
	_ = WhatsAppClient.SendChatPresence(context.Background(), rjid, typeCompose, typeComposeMedia)
}

func WhatsAppSendGPTResponse(event *events.Message, response string) (string, error) {
	if WhatsAppClient != nil {
		var err error

		// Make Sure WhatsApp Client is OK
		if WhatsAppClient.IsConnected() && WhatsAppClient.IsLoggedIn() {
			rJID := event.Info.Chat

			// Compose WhatsApp Proto
			msgExtra := whatsmeow.SendRequestExtra{
				ID: WhatsAppClient.GenerateMessageID(),
			}
			msgContent := &waE2E.Message{
				Conversation: proto.String(response),
			}

			// Send WhatsApp Message Proto
			_, err = WhatsAppClient.SendMessage(context.Background(), rJID, msgContent, msgExtra)
			if err != nil {
				return "", err
			}

			return msgExtra.ID, nil
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

		if bool(WhatsAppGPTTagRegex.MatchString(rMessage)) {
			rMessageSplit := WhatsAppGPTTagRegex.Split(rMessage, 2)

			if len(rMessageSplit) == 2 {
				question := strings.TrimSpace(rMessageSplit[1])

				if len(question) > 0 {
					log.Println(log.LogLevelInfo, "-== Incomming Question ==-")
					log.Println(log.LogLevelInfo, "From     : "+maskRJID)
					log.Println(log.LogLevelInfo, "Question : "+question)

					// Set Chat Presence
					WhatsAppPresence(true)
					WhatsAppComposeStatus(evt.Info.Chat, true, false)
					defer func() {
						WhatsAppComposeStatus(evt.Info.Chat, false, false)
						WhatsAppPresence(false)
					}()

					response, err := gpt.GPTResponse(question)
					if err != nil || len(response) == 0 {
						log.Println(log.LogLevelError, err.Error())
						response = "Sorry, the AI can not response for this time. Please try again after a few moment 🥺"
					}

					_, err = WhatsAppSendGPTResponse(evt, response)
					if err != nil {
						log.Println(log.LogLevelError, "Failed to Send OpenAI GPT Response")
					}
				}
			}
		}
	}
}
