package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

var clientId string = os.Getenv("CLIENT_ID")
var clientSecret string = os.Getenv("CLIENT_SECRET")
var redirectURI string = fmt.Sprintf("%s:%s/auth/callback", os.Getenv("BASE_URL"), os.Getenv("PORT"))

// controller for health check
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Alive!")
}

// controller for rendering the login page
func authHandler(w http.ResponseWriter, r *http.Request) {

	tmpl := template.Must(template.ParseFiles("dist/index.html"))

	data := struct {
		ClientId    string
		Scopes      string
		RedirectURI string
	}{clientId, "user-read-currently-playing", redirectURI}
	tmpl.Execute(w, data)
}

// controller for auth callback
func authCallbackHandler(srv_chan chan<- bool) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		// First, we need to get the value of the `code` query param
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not parse query: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}

		// Get the authorization code
		code := r.FormValue("code")
		auth_error := r.FormValue("error")

		if auth_error != "" {
			log.Println("Error occurred while communicating with Spotify. Make sure you gave the access. ", auth_error)
		}

		spotifyURL := fmt.Sprintf("https://accounts.spotify.com/api/token")

		response, err := http.PostForm(spotifyURL, url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"redirect_uri":  {redirectURI},
			"client_id":     {clientId},
			"client_secret": {clientSecret},
		})

		if err != nil {
			//handle postform error
			log.Println("token error, ", err)
		}

		defer response.Body.Close()

		// Parse the request body into the `oAuthResponse` struct
		var authToken oAuthResponse

		if err := json.NewDecoder(response.Body).Decode(&authToken); err != nil {
			fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}

		// JSONify the authToken
		file, err := json.Marshal(authToken)
		if err != nil {
			log.Println("Unable to parse JSON")
		}

		// Save the access token in a file for future requests
		err = ioutil.WriteFile(os.Getenv("TOKEN_FILE"), file, 0644)
		if err != nil {
			log.Println("Unable to create Token file", err)
		}

		// Render the success page
		tmpl := template.Must(template.ParseFiles("dist/auth_success.html"))
		tmpl.Execute(w, "You have successfully logged in!")

		// Send suceess to server channel to close the server
		srv_chan <- true

	}
}

// For storing OAuth response
type oAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToekn string `json:"refresh_token"`
}