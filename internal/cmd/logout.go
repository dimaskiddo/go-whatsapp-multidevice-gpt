package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/log"
	pkgWhatsApp "github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/whatsapp"
)

// Logout Variable Structure
var Logout = &cobra.Command{
	Use:   "logout",
	Short: "Logout from Go WhatsApp Multi-Device GPT",
	Long:  "Logout from Go WhatsApp Multi-Device GPT",
	Run: func(cmd *cobra.Command, args []string) {
		if pkgWhatsApp.WhatsAppClient != nil {
			pkgWhatsApp.WhatsAppClient.RemoveEventHandlers()
			pkgWhatsApp.WhatsAppLogout()
		}

		log.Println(log.LogLevelInfo, "Successfully Logout from WhatsApp")
	},
}
