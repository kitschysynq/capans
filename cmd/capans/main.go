// Package main provides an ssh server, numbskull
package main

import "github.com/kitschysynq/capans"

func main() {
	// TODO: Configure this bitch
	srv := capans.NewServer()
	srv.Listen(":2222")
}
