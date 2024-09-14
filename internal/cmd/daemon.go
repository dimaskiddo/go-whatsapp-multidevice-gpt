package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	pkgGPT "github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/gpt"
	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/log"
	pkgWhatsApp "github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/whatsapp"
)

// Daemon Variable Structure
var Daemon = &cobra.Command{
	Use:   "daemon",
	Short: "Run as daemon service",
	Long:  "Daemon Service for WhatsApp Multi-Device GPT",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(log.LogLevelInfo, "Go WhatsApp Multi-Device GPT")

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		isHandlerOn := false
		time.Sleep(time.Duration(1) * time.Second)

		for {
			if pkgWhatsApp.WhatsAppClient != nil {
				if pkgWhatsApp.WhatsAppClient.IsConnected() && !isHandlerOn {
					switch strings.ToLower(pkgGPT.WAGPTEngine) {
					case "openai":
						log.Println(log.LogLevelInfo, "Starting WhatsApp Client Event Listener for OpenAI GPT")

					default:
						log.Println(log.LogLevelInfo, "Starting WhatsApp Client Event Listener for Ollama GPT")
					}

					pkgWhatsApp.WhatsAppClient.AddEventHandler(pkgWhatsApp.WhatsAppHandler)

					isHandlerOn = true
				} else if !pkgWhatsApp.WhatsAppClient.IsConnected() {
					log.Println(log.LogLevelWarn, "WhatsApp Client Connection Interupted, Wait 3s for Reloading")

					pkgWhatsApp.WhatsAppClient.RemoveEventHandlers()
					pkgWhatsApp.WhatsAppClient.Disconnect()
					ReloadDatastore()

					time.Sleep(time.Duration(3) * time.Second)
					isHandlerOn = false
				}
			} else {
				log.Println(log.LogLevelWarn, "Waiting for WhatsApp Client to be Logged-in")
				ReloadDatastore()

				isHandlerOn = false
			}

			select {
			case <-sig:
				fmt.Println("")

				if pkgWhatsApp.WhatsAppClient != nil {
					pkgWhatsApp.WhatsAppClient.RemoveEventHandlers()
					pkgWhatsApp.WhatsAppClient.Disconnect()
				}

				log.Println(log.LogLevelInfo, "Terminating Process")
				os.Exit(0)
			case <-time.After(5 * time.Second):
			}
		}
	},
}
