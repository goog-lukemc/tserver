// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
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
	Server  *http.Server
	MUX     *http.ServeMux
	CFG     *ServerConfig
	Brokers sync.Map
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
	if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	return nil
}

// AddBroker adds a broker for later use. The call will return an error
// if the key already exists.
func (s *ServerControl) AddBroker(key interface{}, value interface{}) error {
	if _, ok := s.Brokers.LoadOrStore(key, value); ok {
		return fmt.Errorf("errDuplicateBroker:A broker with this key:%+v already exists, Please use a different key", key)
	}
	return nil
}

func (s *ServerControl) GetBroker(key interface{}) interface{} {
	value, ok := s.Brokers.Load(key)
	if !ok {
		return fmt.Errorf("errBrokerNotFound: Broker with key %v not found", key)
	}
	return value
}
