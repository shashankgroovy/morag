package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/shashankgroovy/morag/server"
	"github.com/shashankgroovy/morag/utils"
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

var baseURI string = os.Getenv("BASE_URI")
var serverPort string = os.Getenv("PORT")

func init() {
	rootCmd.AddCommand(loginCmd)
}

// loginFunc helps authenticate a user by spawning a small server and
// redirecting the user for login process on the browser.
func loginFunc(cmd *cobra.Command, args []string) {

	// check if a user is already authenticated
	if _, err := utils.TestAndSetToken(); err == nil {
		fmt.Println("Use `morag help fetch` to learn more about how to get tracks")
	} else {
		// Spawn a server to initiate the OAuth2 authentication process

		// Initialize a channel for communication with handlers
		srvChan := make(chan bool, 1)

		// Initialize a simple server
		srv := server.App{}
		srv.Initialize(srvChan)

		// Create a goroutine that will open the default browser for authentication
		// as soon as the server is up and running.
		go func() {
			baseURL := fmt.Sprintf("%s:%s", baseURI, serverPort)
			authURL := baseURL + "/auth"

			for {
				time.Sleep(time.Second)

				fmt.Println("You will be shortly redirected to your default browser...")
				resp, err := http.Get(baseURL)
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
			color.Green(authURL)

			// Open the URL in default browser for authentication
			utils.OpenInBrowser(authURL)

			// Wait for a signal to close the server
			<-srvChan
			srv.Shutdown()
			color.Yellow("\nSuccessfully authenticated!")
			fmt.Println("Use `morag help fetch` to learn more about how to get track info from Spotify")
		}()

		// run the server
		srv.Run(os.Getenv("PORT"))
	}
}
