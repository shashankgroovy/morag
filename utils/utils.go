package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fatih/color"
)

var clientId string = os.Getenv("CLIENT_ID")
var clientSecret string = os.Getenv("CLIENT_SECRET")

// OAuthToken struct used for working with OAuth token
type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ValidateAccessToken checks if access-token has expired or not by hitting Spotify
func (token *OAuthToken) ValidateAccessToken() error {
	client := &http.Client{}

	spotifyURL := fmt.Sprintf("https://api.spotify.com/v1/me")
	req, _ := http.NewRequest("GET", spotifyURL, nil)
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)

	response, err := client.Do(req)
	// check if everything's ok
	if err != nil {
		log.Println("Error while validating token", err.Error())
		return err
	}

	if response.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(response.Body)
		fmt.Println("Access token expired")
		return errors.New(string(body))
	}
	defer response.Body.Close()

	return nil
}

// GetNewAccessToken returns a new access token and also writes the new tokens in TOKEN_FILE
func (token *OAuthToken) GetNewAccessToken() error {
	client := &http.Client{}

	spotifyURL := fmt.Sprintf("https://accounts.spotify.com/api/token")
	formData := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {token.RefreshToken},
	}
	// create a post form request
	req, _ := http.NewRequest("POST", spotifyURL, strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientId, clientSecret)

	// fire away
	response, _ := client.Do(req)

	// Persist the new token from http response
	if err := token.SaveTokenToFile(response, true); err != nil {
		fmt.Println(err.Error())
	}

	return nil
}

// SaveTokenToFile retrieves access-token and refresh-token from http.Response
// and persists them to TOKEN_FILE
func (token *OAuthToken) SaveTokenToFile(r *http.Response, refreshed bool) error {

	fmt.Println("Fetching a new access token")

	if !refreshed {
		// Parse the request body into the `OAuthToken` struct
		// This works with brand new tokens.
		err := json.NewDecoder(r.Body).Decode(token)
		if err != nil {
			log.Println("Could not parse JSON response. ", err.Error())
			return err
		}
	} else {
		// While fetching a new access token using refresh token, the refresh token
		// is sometimes not part of the response. And hence we'll use a
		// interface{} to get those values
		var result map[string]interface{}

		if r.StatusCode != http.StatusOK {
			log.Println("Unable to fetch new access token")
			return errors.New("Unable to get new access token")
		}

		err := json.NewDecoder(r.Body).Decode(&result)
		if err != nil {
			log.Println("Could not parse JSON response. ", err.Error())
			return err

		}
		// Replace the access token in token with the new one
		accessToken, ok := result["access_token"].(string)
		if ok {
			token.AccessToken = accessToken
		}
	}

	// JSONify the authToken
	file, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		log.Println("Unable to parse JSON", err.Error())
		return err
	}

	// Save the access token in a file for future requests
	err = ioutil.WriteFile(os.Getenv("TOKEN_FILE"), file, 0644)
	if err != nil {
		log.Println("Unable to create Token file", err.Error())
		return err
	}

	return nil
}

// TestAndSetToken returns an OAuthToken if it can be retrieved from TOKEN_FILE
func TestAndSetToken() (OAuthToken, error) {
	// Initialize the authToken
	var authToken OAuthToken

	tokenFile := os.Getenv("TOKEN_FILE")

	// Check if file exists
	if _, err := os.Stat(tokenFile); err == nil {
		// Token file exists
		Banner()

		// Validate if the file has valid json
		authJson, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			log.Println("Bad json in token file", err.Error())
			log.Println("Use the login command to authenticate again")
			return authToken, err
		}

		// we unmarshal our byteArray which contains our
		// jsonFile's content into 'users' which we defined above
		json.Unmarshal(authJson, &authToken)

		// Check if access token has expired
		err = authToken.ValidateAccessToken()
		if err != nil {
			// Get a new access token
			if err = authToken.GetNewAccessToken(); err == nil {
				fmt.Println("Successfully authenticated!")
			}
		} else {
			fmt.Println("You are authenticated!")
		}
		return authToken, nil
	} else {
		return authToken, err
	}
}

// OpenInBrowser opens a given url in the default browser based on each platform
func OpenInBrowser(url string) error {
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

// Banner prints out the name of cli tool
func Banner() {
	color.Green("")
	color.Green("    __  _______  ____  ___   ______")
	color.Green("   /  |/  / __ \\/ __ \\/   | / ____/")
	color.Green("  / /|_/ / / / / /_/ / /| |/ / __ ")
	color.Green(" / /  / / /_/ / _, _/ ___ / /_/ /")
	color.Green("/_/  /_/\\____/_/ |_/_/  |_\\____/")
	color.Green("")
}
