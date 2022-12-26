package cmd

import (
	"os"
	"strconv"
	"time"

	qrTerm "github.com/mdp/qrterminal/v3"
	"github.com/spf13/cobra"

	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/log"
	pkgWhatsApp "github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/whatsapp"
)

// Login Variable Structure
var Login = &cobra.Command{
	Use:   "login",
	Short: "Login to Go WhatsApp Multi-Device GPT",
	Long:  "Login to Go WhatsApp Multi-Device GPT",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(log.LogLevelInfo, "Go WhatsApp Multi-Device GPT")

		ReloadDatastore()
		if pkgWhatsApp.WhatsAppClient == nil {
			pkgWhatsApp.WhatsAppInitClient(nil)

			qrResponse, qrTimeout, err := pkgWhatsApp.WhatsAppLogin()
			if err != nil {
				log.Println(log.LogLevelError, err.Error())
				return
			}

			if qrResponse == "WhatsApp Client is Reconnected" {
				log.Println(log.LogLevelInfo, qrResponse)
				return
			}

			go func() {
				qrTerm.Generate(qrResponse, qrTerm.L, os.Stdout)
				log.Println(log.LogLevelInfo, "QR Code Will Be Expired in "+strconv.Itoa(qrTimeout)+"s")
			}()

			time.Sleep(time.Duration(qrTimeout) * time.Second)
		} else {
			log.Println(log.LogLevelInfo, "WhatsApp Client Already Logged-in")
		}
	},
}
