package server

import (
	"flag"
	"net/http"

	"github.com/gorilla/mux"
)

// configureRoutes setups up app routes and static routes
func configureRoutes(srvChan chan<- bool) *mux.Router {

	r := mux.NewRouter().StrictSlash(true)
	// Serve static files
	r = setupStaticRoutes(r)

	// Health check
	r.HandleFunc("/", healthCheckHandler).Methods(http.MethodGet)
	r.HandleFunc("/health", healthCheckHandler).Methods(http.MethodGet)
	r.HandleFunc("/auth", authHandler).Methods(http.MethodGet)

	// create a function closure for authCallbackHandler to work with server
	// close channel
	authCallbackHandlerWithChannel := authCallbackHandler(srvChan)
	r.HandleFunc("/auth/callback", authCallbackHandlerWithChannel).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)

	return r
}

// Creates routes for serving static assets
func setupStaticRoutes(r *mux.Router) *mux.Router {

	// Serve static files
	var dir string
	flag.StringVar(&dir, "dir", "dist", "The directory for static file content")
	flag.Parse()
	r.PathPrefix("/dist/").Handler(http.StripPrefix(
		"/dist/",
		http.FileServer(http.Dir(dir))),
	)

	return r

}
