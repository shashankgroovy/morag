package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

// configureRoutes setups up app routes and static routes
func configureRoutes() *mux.Router {

	r := mux.NewRouter().StrictSlash(true)

	// Health check
	r.HandleFunc("/", healthCheckHandler).Methods(http.MethodGet)
	r.HandleFunc("/health", healthCheckHandler).Methods(http.MethodGet)

	r.HandleFunc("/auth", authHandler).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/auth/callback", authCallbackHandler).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)

	return r
}
