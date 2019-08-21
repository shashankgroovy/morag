package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/shashankgroovy/morag/utils"
	"github.com/spf13/cobra"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetches track information for an artist.",
	Long: `Fetch helps you download the entire catalog/library of an artist (of
your choice) from Spotify and save that in an output.csv file. A valid
artistID or multiple artistIDs can be passed separated by space as
arguments to this command.

USAGE:
$ morag fetch [artistID]

EXAMPLE:
$ morag fetch 0OdUWJ0sBjDrqHygGUXeCF
`,
	Run: fetch,
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	// Add a local flag which will only run when this command
	// is called directly.
	fetchCmd.Flags().StringP("output_csv", "o", "output.csv", "Provide an output file name of your choice")
}

func fetch(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		// Print error, help text and exit
		fmt.Printf("\nERROR: Please provide a Spotify artistID.\n\n")
		cmd.Help()
	} else {
		artistID := args[0]

		// check if a user is already authenticated
		if authToken, err := utils.TestAndSetToken(); err != nil {
			log.Println("Error while setting the auth token", err.Error())
		} else {
			client := &http.Client{}

			spotifyURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%s/albums", artistID)
			req, _ := http.NewRequest("GET", spotifyURL, nil)
			req.Header.Add("Authorization", "Bearer "+authToken.AccessToken)

			response, err := client.Do(req)
			// check if everything's ok
			body, _ := ioutil.ReadAll(response.Body)
			if err != nil || response.StatusCode != http.StatusOK {
				log.Println("Unable to fetch new access token", err.Error(), string(body))
			}
			fmt.Println("Fetching albums for artist")
			fmt.Println(string(body))
			defer response.Body.Close()
		}
	}
}
