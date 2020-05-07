package epserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	StaticDir    string
}

type ServerControl struct {
	Server *http.Server
	MUX    *http.ServeMux
	CFG    *ServerConfig
}

// Start the http listener as configured
func (s *ServerControl) Start(fs ...func(*ServerControl)) {
	// Create a shutdown signal channel to handle shutdowns gracefully
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	// Load the handlers passed
	for _, f := range fs {
		f(s)
	}

	// Start the listener
	go func() {
		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("errHTTPServer:%s", err)
		}
	}()

	// Our server will block here and will stop until there is an OS interupt
	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Shutdown the server with a timeout
	if err := s.Server.Shutdown(ctx); err != nil {
		log.Printf("errServerShutdown:%s", err)
	}
	log.Printf("Server Exiting")
}

// NewServer create a new HTTP endpoint ready to start
func NewServer(cfg *ServerConfig) *ServerControl {

	mux := http.NewServeMux()
	// Create the server
	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	return &ServerControl{
		Server: server,
		MUX:    mux,
		CFG:    cfg,
	}
}

// GetRequestBody decode the body of the server to the target type
func GetRequestBody(w http.ResponseWriter, r *http.Request, target *interface{}) error {
	// Check if the the http method is a post request.
	if r.Method != "POST" {
		http.Error(w, "errBadHTTPMethod", http.StatusMethodNotAllowed)
		return fmt.Errorf("errBadHTTPMethod")
	}

	// Decode the http body to a dlp request
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	return nil
}
