// Package server provides HTTP server implementation.
package server

// Server defines the interface for a server.
type Server interface {
	// Run starts the server and blocks until interrupted.
	Run()
	// Start starts the server in a goroutine.
	Start()
	// Stop gracefully stops the server.
	Stop()
}