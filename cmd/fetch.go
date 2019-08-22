package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/mitchellh/mapstructure"
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
		var MAX_LIMIT = 50

		// check if a user is already authenticated
		if authToken, err := utils.TestAndSetToken(); err != nil {
			log.Println("Error while setting the auth token", err.Error())
		} else {
			// Create an empty interface to hold all albums
			album_ch := make(chan string)
			//track_ch := make(chan string)

			go getAlbums(authToken, artistID, album_ch, 0, MAX_LIMIT)

			for albumId := range album_ch {
				go getAlbumTracks(albumId)
			}
			close(album_ch)
			fmt.Println("Reached here")
		}
	}
}

// getAlbums returns the result of getting an album
func getAlbums(authToken utils.OAuthToken, artistID string, album_ch chan<- string, offset, limit int) {
	var (
		albums []interface{}
		result map[string]interface{}
	)

	// Create a new http client
	client := &http.Client{}

	// Construct the http request
	spotifyURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%s/albums", artistID)
	req, _ := http.NewRequest("GET", spotifyURL, nil)
	req.Header.Add("Authorization", "Bearer "+authToken.AccessToken)

	// Add the pagination query parameters
	q := req.URL.Query()
	q.Add("offset", string(fmt.Sprintf("%d", offset)))
	q.Add("limit", string(fmt.Sprintf("%d", limit)))
	req.URL.RawQuery = q.Encode()

	fmt.Println(req.URL.String())

	// Fire it away
	fmt.Println("Fetching albums of artist")
	resp, err := client.Do(req)

	// check if everything's ok
	if err != nil {
		log.Println("Error in request", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("404 ", resp.StatusCode, string(body))
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Println("Could not parse JSON response. ", err.Error())

	}
	// Store all albums from request
	albums = result["items"].([]interface{})
	for _, value := range albums {
		var album utils.Album
		mapstructure.Decode(value, &album)
		album_ch <- album.Id
	}
}

// getAlbumTracks fetches all the tracks of an album
func getAlbumTracks(albumId string) {
	fmt.Println(albumId)
}
