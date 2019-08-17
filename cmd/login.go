package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login connects you to your Spotify account.",
	Long: `Login authenticates you to your Spotify account using OAuth2.
It opens up a login page in the default browser to connect your Spotify account and
obtain an access token.

Simply issue: "morag login" to initiate the authentication process.`,
	Run: loginFunc,
}

func init() {
	rootCmd.AddCommand(loginCmd)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func loginFunc(cmd *cobra.Command, args []string) {
	fmt.Println("Login called")
	open("http://www.google.com")
}

// Opens url in the default browser
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
