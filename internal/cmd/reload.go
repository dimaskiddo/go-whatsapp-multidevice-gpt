package cmd

import (
	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/log"
	pkgWhatsApp "github.com/dimaskiddo/go-whatsapp-multidevice-gpt/pkg/whatsapp"
)

func ReloadDatastore() {
	pkgWhatsApp.WhatsAppClient = nil

	devices, err := pkgWhatsApp.WhatsAppDatastore.GetAllDevices()
	if err != nil {
		log.Println(log.LogLevelError, "Failed to Load WhatsApp Client Devices from Datastore")
	}

	for _, device := range devices {
		realJID := device.ID.User
		maskJID := realJID[0:len(realJID)-4] + "xxxx"

		log.Println(log.LogLevelInfo, "Restoring WhatsApp Client Connection for "+maskJID)
		pkgWhatsApp.WhatsAppInitClient(device)

		err = pkgWhatsApp.WhatsAppReconnect()
		if err != nil {
			log.Println(log.LogLevelError, err.Error())
		}
	}
}
