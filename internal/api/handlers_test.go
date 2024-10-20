package api

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	r.DELETE("/repositories/:owner/:repo", DeleteRepository)
	r.GET("/repositories/:owner/:repo/pulls", ListOpenPullRequests)
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
	t.Run("GitHub API Failure", func(t *testing.T) {
		httpmock.RegisterResponder("POST", "https://api.github.com/user/repos",
			httpmock.NewStringResponder(http.StatusInternalServerError, `{"error": "failed to create repository: status code 500"}`))

		repo := map[string]string{
			"name": "test-repo",
		}
		jsonValue, _ := json.Marshal(repo)

		req, _ := http.NewRequest(http.MethodPost, "/repositories", bytes.NewBuffer(jsonValue))
		req.Header.Set("Authorization", "Bearer placeholder")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code) //erro 500

		var errorResponse map[string]string
		json.Unmarshal(resp.Body.Bytes(), &errorResponse)
		assert.Equal(t, "failed to create repository: status code 500", errorResponse["error"])
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

func TestDeleteRepository(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)

	t.Run("Successful Repository Deletion", func(t *testing.T) {

		httpmock.RegisterResponder("DELETE", "https://api.github.com/repos/testuser/test-repo",
			httpmock.NewStringResponder(http.StatusNoContent, ``))

		req, _ := http.NewRequest(http.MethodDelete, "/repositories/testuser/test-repo", nil)
		req.Header.Set("Authorization", "Bearer your_github_token")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusNoContent, resp.Code)
	})

	t.Run("Unauthorized (Missing Token)", func(t *testing.T) {
		owner := "testuser"
		repoName := "test-repo"

		req, _ := http.NewRequest(http.MethodDelete, "/repositories/"+owner+"/"+repoName, nil)

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("GitHub API Error", func(t *testing.T) {
		owner := "testuser"
		repoName := "test-repo"

		httpmock.RegisterResponder("DELETE", "https://api.github.com/repos/"+owner+"/"+repoName,
			httpmock.NewStringResponder(http.StatusInternalServerError, `{"error": "not found"}`))

		req, _ := http.NewRequest(http.MethodDelete, "/repositories/"+owner+"/"+repoName, nil)
		req.Header.Set("Authorization", "Bearer placeholder")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("Deletion with Incorrect Repo", func(t *testing.T) {
		owner := "wronguser"
		repoName := "wrong-repo"

		httpmock.RegisterResponder("DELETE", "https://api.github.com/repos/"+owner+"/"+repoName,
			httpmock.NewStringResponder(http.StatusNotFound, `{"error": "not found"}`))

		req, _ := http.NewRequest(http.MethodDelete, "/repositories/"+owner+"/"+repoName, nil)
		req.Header.Set("Authorization", "Bearer your_github_token")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})
}

func TestListOpenPullRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)
	mock_repo := RepoPullRe{
		Name:  "mock-repo",
		Owner: "mock-user",
		PullRequests: []PullRequest{
			{Title: "Fix bug in feature X", Number: 1, State: "open"},
			{Title: "Add new API endpoint", Number: 2, State: "open"},
			{Title: "Improve documentation", Number: 3, State: "closed"},
		},
	}

	t.Run("Sucessfully retrieving the number of pull requests", func(t *testing.T) {

		pullRequestsJSON, err := json.MarshalIndent(mock_repo.PullRequests, "", "  ")
		if err != nil {
			fmt.Println("Error converting to JSON:", err)
			return
		}
		httpmock.RegisterResponder("GET", "https://api.github.com/repos/testuser/test-repo/pulls",
			httpmock.NewStringResponder(http.StatusOK, string(pullRequestsJSON)))

		req, _ := http.NewRequest(http.MethodGet, "/repositories/testuser/test-repo/pulls", nil)
		req.Header.Set("Authorization", "Bearer placeholder")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var response map[string]int
		json.Unmarshal(resp.Body.Bytes(), &response)
		assert.Equal(t, 3, response["open_pull_requests"])
	})
	t.Run("Unauthorized Request (Missing Token)", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/repositories/testuser/test-repo/pulls", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("GitHub API Error", func(t *testing.T) {
		httpmock.RegisterResponder("GET", "https://api.github.com/repos/testuser/test-repo/pulls?state=open",
			httpmock.NewStringResponder(http.StatusInternalServerError, `{"message": "Internal Server Error"}`))

		req, _ := http.NewRequest(http.MethodGet, "/repositories/testuser/test-repo/pulls", nil)
		req.Header.Set("Authorization", "Bearer placeholder")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)

		var response map[string]string
		json.Unmarshal(resp.Body.Bytes(), &response)

		assert.Equal(t, "failed to get pull requests: status code 500", response["error"])
	})
}
