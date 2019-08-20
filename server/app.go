package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// App struct for keeping things simple and concise.
type App struct {
	Router *mux.Router
	Server *http.Server
}

// Initialize sets up routing
func (a *App) Initialize(srvChan chan<- bool) {

	// Setup routes
	a.Router = configureRoutes(srvChan)
}

// Run starts an http.Server
func (a *App) Run(httpPort string) {
	// Setup server
	a.Server = &http.Server{
		Handler:      a.Router,
		Addr:         ":" + httpPort,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	a.Server.ListenAndServe()
}

// Shutdown safely closes the http.Server
func (a *App) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.Server.Shutdown(ctx); err != nil {
		// handle err
		log.Println("Couldn't shutdown", err)
	}

}
