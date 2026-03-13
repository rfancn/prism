// Package main is the entry point of the Prism HTTP/HTTPS request relay service.
package main

import "github.com/rfancn/prism/cmd"

//go:generate sqlc generate -f assets/sqlc.yaml
func main() {
	cmd.Execute()
}