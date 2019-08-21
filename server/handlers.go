package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/shashankgroovy/morag/utils"
)

var clientId string = os.Getenv("CLIENT_ID")
var clientSecret string = os.Getenv("CLIENT_SECRET")
var redirectURI string = fmt.Sprintf("%s:%s/auth/callback", os.Getenv("BASE_URI"), os.Getenv("PORT"))

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
func authCallbackHandler(srvChan chan<- bool) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		// First, we need to get the value of the `code` query param
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not parse query: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}

		// Get the authorization code
		code := r.FormValue("code")
		authError := r.FormValue("error")

		if authError != "" {
			log.Println("Error occurred while communicating with Spotify. Make sure you gave the access. ", authError)
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

		// Parse the request body into the `OAuthToken` struct
		var authToken utils.OAuthToken

		if err := authToken.SaveTokenToFile(response, false); err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		// Render the success page
		tmpl := template.Must(template.ParseFiles("dist/auth_success.html"))
		tmpl.Execute(w, "You have successfully logged in!")

		// Send suceess to server channel to close the server
		srvChan <- true

	}
}
