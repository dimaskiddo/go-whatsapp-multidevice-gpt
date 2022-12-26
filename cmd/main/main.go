package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dimaskiddo/go-whatsapp-multidevice-gpt/internal/cmd"
)

// Root Variable Structure
var r = &cobra.Command{
	Use:   "go-whatsapp-multidevice-gpt",
	Short: "Go WhatsApp Multi-Device GPT",
	Long:  "Go WhatsApp Multi-Device GPT",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Init Function
func init() {
	// Initialize Command
	r.AddCommand(cmd.Version)
	r.AddCommand(cmd.Daemon)
	r.AddCommand(cmd.Login)
	r.AddCommand(cmd.Logout)
}

// Main Function
func main() {
	err := r.Execute()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
