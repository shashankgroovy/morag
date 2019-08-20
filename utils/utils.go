package utils

import (
	"encoding/json"
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

// For working with OAuth token
type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Checks if access-token has expired or not by hitting Spotify
func (token *OAuthToken) ValidateAccessToken() error {
	client := &http.Client{}

	spotifyURL := fmt.Sprintf("https://api.spotify.com/v1/me")
	req, _ := http.NewRequest("GET", spotifyURL, nil)
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)

	response, err := client.Do(req)
	defer response.Body.Close()

	// check if everything's ok
	if err != nil || response.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(response.Body)
		log.Println("Unable to fetch new access token", err.Error(), string(body))
		return err
	}

	return nil
}

// Returns a new access token and also writes the new tokens in TOKEN_FILE
func (token *OAuthToken) GetNewAccessToken() error {
	client := &http.Client{}

	spotifyURL := fmt.Sprintf("https://accounts.spotify.com/api/token")
	formData := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {token.RefreshToken},
	}
	// send a post form request
	req, _ := http.NewRequest("POST", spotifyURL, strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientId, clientSecret)

	// Base64 encode client id and client secret for basic auth
	// clientCredString := fmt.Sprintf("%s:%s", clientId, clientSecret)
	// encodedClientCred := base64.StdEncoding.EncodeToString([]byte(clientCredString))
	// req.Header.Add("Authorization", "Basic "+encodedClientCred)

	response, _ := client.Do(req)
	body, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Println("Unable to fetch new access token", string(body))
	}

	return nil
}

// SaveTokenToFile retrieves access-token and refresh-token from http.Response
// and persists them to TOKEN_FILE
func (token *OAuthToken) SaveTokenToFile(r *http.Response) error {

	// Parse the request body into the `OAuthToken` struct
	if err := json.NewDecoder(r.Body).Decode(token); err != nil {
		log.Println("Could not parse JSON response. ", err.Error())
		return err
	}

	// JSONify the authToken
	file, err := json.Marshal(token)
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

// Opens url in the default browser
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
