package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/goncalo0domingos/github-api/internal/api"
)

func main() {
	r := gin.Default()

	// Set up the POST /repositories endpoint
	r.POST("/repositories", api.CreateRepository)
	r.GET("/repositories", api.ListRepositories)
	r.DELETE("/repositories/:owner/:repo", api.DeleteRepository)
	r.GET("/repositories/:owner/:repo/pulls", api.ListOpenPullRequests)

	log.Println("Server starting on port 8080...")
	log.Fatal(r.Run(":8080")) // Start the server on port 8080
}
