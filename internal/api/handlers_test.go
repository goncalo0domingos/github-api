package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/jarcoal/httpmock"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/repositories", CreateRepository)
	return r
}

func TestCreateRepository(t *testing.T) {
	gin.SetMode(gin.TestMode) // Set Gin to test mode
	router := setupRouter()

	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)

	t.Run("Valid Repository Creation", func(t *testing.T) {
		httpmock.RegisterResponder("POST", "https://api.github.com/user/repos",
			httpmock.NewStringResponder(http.StatusCreated, ``))
		// Prepare a valid repository creation request
		repo := map[string]string{
			"name": "test-repo",
		}
		jsonValue, _ := json.Marshal(repo)

		// Create a new HTTP request with valid token
		req, _ := http.NewRequest(http.MethodPost, "/repositories", bytes.NewBuffer(jsonValue))
		req.Header.Set("Authorization", "Bearer your_github_token") // Use a test token

		// Record the response
		resp := httptest.NewRecorder()

		// Perform the request
		router.ServeHTTP(resp, req)

		// Assert the response
		assert.Equal(t, http.StatusCreated, resp.Code)

		var createdRepo map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &createdRepo)

		assert.Equal(t, "test-repo", createdRepo["name"])
	})

	t.Run("Invalid Repository Creation (Missing Name)", func(t *testing.T) {
		// Prepare an invalid request (missing name)
		repo := map[string]string{}
		jsonValue, _ := json.Marshal(repo)

		// Create a new HTTP request
		req, _ := http.NewRequest(http.MethodPost, "/repositories", bytes.NewBuffer(jsonValue))
		req.Header.Set("Authorization", "Bearer your_github_token") // Use a test token

		// Record the response
		resp := httptest.NewRecorder()

		// Perform the request
		router.ServeHTTP(resp, req)

		// Assert the response
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Unauthorized (Missing Token)", func(t *testing.T) {
		// Prepare a valid repository creation request
		repo := map[string]string{
			"name": "test-repo",
		}
		jsonValue, _ := json.Marshal(repo)

		// Create a new HTTP request without an Authorization header
		req, _ := http.NewRequest(http.MethodPost, "/repositories", bytes.NewBuffer(jsonValue))

		// Record the response
		resp := httptest.NewRecorder()

		// Perform the request
		router.ServeHTTP(resp, req)

		// Assert the response
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}
