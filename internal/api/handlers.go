package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

var repositories []Repository

func CreateRepository(c *gin.Context) {
	var repo Repository
	if err := c.ShouldBindJSON(&repo); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	repositories = append(repositories, repo)
	c.JSON(201, repo)
}

func ListRepositories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repositories)
}
