package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/repositories", CreateRepository)
	r.GET("/repositories", ListRepositories)
	return r
}

func TestCreateRepository(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)

	t.Run("Valid Repository Creation", func(t *testing.T) {
		httpmock.RegisterResponder("POST", "https://api.github.com/user/repos",
			httpmock.NewStringResponder(http.StatusCreated, ``))

		repo := map[string]string{
			"name": "test-repo",
		}
		jsonValue, _ := json.Marshal(repo)

		req, _ := http.NewRequest(http.MethodPost, "/repositories", bytes.NewBuffer(jsonValue))
		req.Header.Set("Authorization", "Bearer placeholder")

		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Code)

		var createdRepo map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &createdRepo)

		assert.Equal(t, "test-repo", createdRepo["name"])
	})

	t.Run("Invalid Repository Creation (Missing Name)", func(t *testing.T) {
		repo := map[string]string{}
		jsonValue, _ := json.Marshal(repo)

		req, _ := http.NewRequest(http.MethodPost, "/repositories", bytes.NewBuffer(jsonValue))
		req.Header.Set("Authorization", "Bearer placeholder")

		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Unauthorized (Missing Token)", func(t *testing.T) {
		repo := map[string]string{
			"name": "test-repo",
		}
		jsonValue, _ := json.Marshal(repo)

		req, _ := http.NewRequest(http.MethodPost, "/repositories", bytes.NewBuffer(jsonValue))

		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}

func TestListRepositories(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)

	t.Run("Valid List Repositories", func(t *testing.T) {
		mockRepos := `[
            {
                "name": "repo1",
                "private": false
            },
            {
                "name": "repo2",
                "private": true
            }
        ]`
		httpmock.RegisterResponder("GET", "https://api.github.com/user/repos",
			httpmock.NewStringResponder(http.StatusOK, mockRepos))

		req, _ := http.NewRequest(http.MethodGet, "/repositories", nil)
		req.Header.Set("Authorization", "Bearer placeholder")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code) // erro 200

		var responseRepo []RepoPrint
		err := json.Unmarshal(resp.Body.Bytes(), &responseRepo)
		assert.Nil(t, err)

		assert.Len(t, responseRepo, 2)
		assert.Equal(t, "repo1", responseRepo[0].Name)
		assert.False(t, responseRepo[0].Private)
		assert.Equal(t, "repo2", responseRepo[1].Name)
		assert.True(t, responseRepo[1].Private)
	})

	t.Run("Unauthorized Request", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/repositories", nil)

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)

		var errorResponse map[string]string
		json.Unmarshal(resp.Body.Bytes(), &errorResponse)
		assert.Equal(t, "Authorization token required", errorResponse["error"])
	})

	t.Run("GitHub API Failure", func(t *testing.T) {
		httpmock.RegisterResponder("GET", "https://api.github.com/user/repos",
			httpmock.NewStringResponder(http.StatusInternalServerError, `{"message": "Internal Server Error"}`))

		req, _ := http.NewRequest(http.MethodGet, "/repositories", nil)
		req.Header.Set("Authorization", "Bearer placeholder")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code) //erro 500

		var errorResponse map[string]string
		json.Unmarshal(resp.Body.Bytes(), &errorResponse)
		assert.Equal(t, "failed to get repositories: status code 500", errorResponse["error"])
	})
}
