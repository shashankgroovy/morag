package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/mitchellh/mapstructure"
	"github.com/mohae/struct2csv"
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
	// fetchCmd.Flags().StringP("output_csv", "o", "output.csv", "Provide an output file name of your choice")
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

			// create some channels for data exchange
			albumCh := make(chan string)
			trackCh := make(chan string)
			// retry channel to stop creating more goroutines as soon as a rate limit is hit
			retryCh := make(chan time.Duration)

			var songlist []utils.FullSoundtrack
			var mu sync.Mutex

			var wg sync.WaitGroup

			// paginate this
			go getAlbums(authToken, artistID, albumCh, 0, MAX_LIMIT)

			albumNum := 0

			for albumId := range albumCh {
				select {
				case sleep := <-retryCh:
					// Pause go routine creation untill cooldown
					time.Sleep(sleep * time.Millisecond)
				default:
					// do nothing
				}
				wg.Add(1)
				albumNum += 1
				color.Yellow("[albumNum]" + string(fmt.Sprintf("%d", albumNum)))
				go getAlbumTracks(authToken, albumId, trackCh, retryCh, &wg)
			}

			trackNum := 0
			for trackId := range trackCh {
				select {
				case sleep := <-retryCh:
					// Pause go routine creation untill cooldown
					// Time to write things to csv
					WritetoCSV(songlist)

					time.Sleep(sleep * time.Millisecond)
				default:
					// do nothing
				}
				trackNum += 1
				color.Yellow("[trackNum]" + string(fmt.Sprintf("%d", trackNum)))
				wg.Add(1)
				go getFullSoundTrack(authToken, trackId, &songlist, retryCh, &wg, &mu)
			}

			wg.Wait()

			WritetoCSV(songlist)

			fmt.Println("Finished")
			fmt.Println("Output stored at - ", os.Getenv("OUTPUT_FILE"))
		}
	}
}

// getAlbums returns the result of getting an album
func getAlbums(authToken utils.OAuthToken, artistID string, albumCh chan<- string, offset, limit int) {
	color.Yellow("[getAlbums] get albums")
	defer close(albumCh)

	var (
		albums []interface{}
		result map[string]interface{}
	)

	// Create a new http client
	client := &http.Client{}
	// a basic flag to safe guard process execution. It's helpful when rate limit is hit
	proceed := true

	// Construct the http request
	spotifyURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%s/albums", artistID)
	req, _ := http.NewRequest("GET", spotifyURL, nil)
	req.Header.Add("Authorization", "Bearer "+authToken.AccessToken)

	// Add the pagination query parameters
	q := req.URL.Query()
	q.Add("offset", string(fmt.Sprintf("%d", offset)))
	q.Add("limit", string(fmt.Sprintf("%d", limit)))
	req.URL.RawQuery = q.Encode()

	// Fire it away
	color.Yellow("[getAlbums] Fetching albums of artist")
	fmt.Println(req.URL.String())
	resp, err := client.Do(req)

	// check if everything's ok
	if err != nil {
		log.Println("Error in request", err.Error())
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		proceed = false
		log.Println("[getAlbums] 429", resp.StatusCode, resp.Header.Get("Retry-After"))
		retryAfter, _ := strconv.Atoi(resp.Header.Get("Retry-After"))

		// Start the retry mechanism
		retry := utils.RetryRequest{Attempt: 1, Min: 1, Max: 5}
		retry.Backoff(retryAfter)

		//Execute this request again
		for retry.Attempt < retry.Max {
			retry.Attempt += 1
			time.Sleep(retry.Duration * time.Second)

			// fire the request
			resp, err = client.Do(req)
			if resp.StatusCode != http.StatusTooManyRequests {
				proceed = true
				break
			}
		}
	} else if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Error ", resp.StatusCode, string(body))
	}
	defer resp.Body.Close()

	if proceed {
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			log.Println("Could not parse JSON response. ", err.Error())

		}
		// Store all albums from request
		albums = result["items"].([]interface{})
		for _, value := range albums {
			var album utils.SimplifiedAlbum
			mapstructure.Decode(value, &album)
			albumCh <- album.Id
		}
	}
}

// getAlbumTracks fetches all the tracks of an album
func getAlbumTracks(authToken utils.OAuthToken, albumId string, trackCh chan<- string, retryCh chan<- time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	var (
		tracks []interface{}
		result map[string]interface{}
	)

	offset := 0
	MAX_LIMIT := 50

	// Create a new http client
	client := &http.Client{}

	// a basic flag to safe guard process execution. It's helpful when rate limit is hit
	proceed := true

	// Construct the http request
	spotifyURL := fmt.Sprintf("https://api.spotify.com/v1/albums/%s/tracks", albumId)
	req, _ := http.NewRequest("GET", spotifyURL, nil)
	req.Header.Add("Authorization", "Bearer "+authToken.AccessToken)

	// Add the pagination query parameters
	q := req.URL.Query()
	q.Add("offset", string(fmt.Sprintf("%d", offset)))
	q.Add("limit", string(fmt.Sprintf("%d", MAX_LIMIT)))
	req.URL.RawQuery = q.Encode()

	// Fire it away
	color.Cyan("\n[getAlbumTracks] Getting tracks for")
	fmt.Println(req.URL.String())
	resp, err := client.Do(req)

	// check if everything's ok
	if err != nil {
		log.Println("Error in request", err.Error())
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		proceed = false
		log.Println("[getAlbumTracks] 429", resp.StatusCode, resp.Header.Get("Retry-After"))
		retryAfter, _ := strconv.Atoi(resp.Header.Get("Retry-After"))

		// Start the retry mechanism
		retry := utils.RetryRequest{Attempt: 1, Min: 1, Max: 5}
		retry.Backoff(retryAfter)
		retryCh <- retry.Duration

		//Execute this request again
		for retry.Attempt < retry.Max {
			retry.Attempt += 1
			time.Sleep(retry.Duration * time.Second)

			// fire the request
			resp, err = client.Do(req)
			if resp.StatusCode != http.StatusTooManyRequests {
				proceed = true
				break
			}
		}
	} else if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Error ", resp.StatusCode, string(body))
	}
	defer resp.Body.Close()

	if proceed {
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			log.Println("Could not parse JSON response. ", err.Error())

		}
		// Store all tracks from request
		tracks = result["items"].([]interface{})
		for _, value := range tracks {
			var soundtrack utils.SimplifiedSoundtrack
			mapstructure.Decode(value, &soundtrack)
			trackCh <- soundtrack.Id
		}

	}
}

// getFullSoundTrack retrieves a list of full soundtracks
func getFullSoundTrack(authToken utils.OAuthToken, trackId string, songlist *[]utils.FullSoundtrack, retryCh chan<- time.Duration, wg *sync.WaitGroup, mu *sync.Mutex) {
	var soundtrack utils.FullSoundtrack

	// Create a new http client
	client := &http.Client{}

	// a basic flag to safe guard process execution. It's helpful when rate limit is hit
	proceed := true

	// Construct the http request
	spotifyURL := fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", trackId)
	req, _ := http.NewRequest("GET", spotifyURL, nil)
	req.Header.Add("Authorization", "Bearer "+authToken.AccessToken)

	// Add the pagination query parameters
	q := req.URL.Query()
	//q.Add("ids", string(fmt.Sprintf("%s", trackList)))
	req.URL.RawQuery = q.Encode()

	// Fire it away
	color.Red("Fetching soundtrack")
	fmt.Println(req.URL.String())
	resp, err := client.Do(req)

	// check if everything's ok
	if err != nil {
		log.Println("Error in request", err.Error())
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		proceed = false
		log.Println("[getFullSoundTrack] 429", resp.StatusCode, resp.Header.Get("Retry-After"))
		retryAfter, _ := strconv.Atoi(resp.Header.Get("Retry-After"))

		// Start the retry mechanism
		retry := utils.RetryRequest{Attempt: 1, Min: 1, Max: 5}
		retry.Backoff(retryAfter)
		retryCh <- retry.Duration

		//Execute this request again
		for retry.Attempt < retry.Max {
			retry.Attempt += 1
			time.Sleep(retry.Duration * time.Second)

			// fire the request
			resp, err = client.Do(req)
			if resp.StatusCode != http.StatusTooManyRequests {
				proceed = true
				break
			}
		}
	} else if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Error ", resp.StatusCode, string(body))
	}
	defer resp.Body.Close()

	if proceed {
		err = json.NewDecoder(resp.Body).Decode(&soundtrack)
		if err != nil {
			log.Println("Could not parse JSON response. ", err.Error())

		}
		mu.Lock()
		*songlist = append(*songlist, soundtrack)
		mu.Unlock()
	}
}

// Writes a song to CSV
func WritetoCSV(songlist []utils.FullSoundtrack) {

	log.Println("[WRITER] Writing to file")

	file, err := os.OpenFile(os.Getenv("OUTPUT_FILE"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatal("[WRITER] Cannot create file", err)
	}
	defer file.Close()

	w := struct2csv.NewWriter(file)
	w.SetComma('\t')
	w.SetSeparators("|", "|")

	err = w.WriteStructs(songlist)
	if err != nil {
		// handle error
		log.Println("[WRITER] Error", err.Error())
	}

}
