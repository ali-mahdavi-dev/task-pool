package main

import (
	"log"

	"task-pool/cmd/command"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

// @title						task-pool API Documentation
// @version					1.0.0
// @description				task-pool API documentation
// @schemes					http https
// @securityDefinitions.apikey	BearerAuth
// @type						apiKey
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and your JWT token.
func main() {
	command.Execute()
}
