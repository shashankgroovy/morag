package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logs out a current user.",
	Long: `Logout removes any previously persisted access tokens to safetly
disconnect from Spotify.

Simple run: "morag logout" to logout.`,
	Run: logout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func logout(cmd *cobra.Command, args []string) {
	tokenFile := os.Getenv("TOKEN_FILE")

	if _, err := os.Stat(tokenFile); err == nil {
		_ = os.Remove(tokenFile)
		fmt.Println("Logged out")

	} else if os.IsNotExist(err) {
		fmt.Println("Already logged out")

	}

}
