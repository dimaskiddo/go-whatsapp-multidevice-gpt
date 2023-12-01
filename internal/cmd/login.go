package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

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

			fmt.Println("")
			fmt.Println("Please Insert Your Phone Number to Generate Pair Code:")
			phoneInput := bufio.NewReader(os.Stdin)
			fmt.Println("")

			phoneNumber, err := phoneInput.ReadString('\n')
			if err != nil {
				log.Println(log.LogLevelError, "Failed to Get Phone Number Input!")
			}

			pairResponse, pairTimeout, err := pkgWhatsApp.WhatsAppLogin(phoneNumber)
			if err != nil {
				log.Println(log.LogLevelError, err.Error())
				return
			}

			if pairResponse == "WhatsApp Client is Reconnected" {
				log.Println(log.LogLevelInfo, pairResponse)
				return
			}

			log.Println(log.LogLevelInfo, "Successfully Generate Pair Code. Your Pair Code is "+pairResponse)
			log.Println(log.LogLevelInfo, "Pair Code Will Be Expired in "+strconv.Itoa(pairTimeout)+"s")
		} else {
			log.Println(log.LogLevelInfo, "WhatsApp Client Already Logged-in")
		}
	},
}
