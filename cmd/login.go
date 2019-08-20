package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/shashankgroovy/morag/server"
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
}

// loginFunc helps authenticate a user by spawning a small server and
// redirecting the user for login process on the browser.
func loginFunc(cmd *cobra.Command, args []string) {

	// Make a channel to communicate with handlers
	srv_chan := make(chan bool, 1)

	// Initialize a simple server
	srv := server.App{}
	srv.Initialize(srv_chan)

	// Create a goroutine that will open the default browser for authentication
	// as soon as the server is up and running.
	go func() {
		base_uri := fmt.Sprintf("http://localhost:%s", os.Getenv("PORT"))
		auth_uri := base_uri + "/auth"

		for {
			time.Sleep(time.Second)

			fmt.Println("You will be shortly redirected to your default browser...")
			resp, err := http.Get(base_uri)
			if err != nil {
				fmt.Println("Failed:", err)
				continue
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				fmt.Println("Not OK:", resp.StatusCode)
				continue
			}

			// Reached this point: server is up and running!
			break
		}
		fmt.Printf("If the browser doesn't open automatically then simply use the following URL:\n\n")
		color.Green(auth_uri)

		// Open the URL in default browser for authentication
		open(auth_uri)

		// Wait for a signal to close the server
		<-srv_chan
		srv.Shutdown()
		color.Yellow("\nSuccessfully authenticated!")
		fmt.Println("You can now use the `fetch` command to get tracks from Spotify")
	}()

	// run the server
	srv.Run(os.Getenv("PORT"))
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
