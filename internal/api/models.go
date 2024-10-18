package api

// repo variables
type Repository struct {
	Name string `json:"name" binding:"required"`
}

type RepoPrint struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
}

type PullRequest struct {
	Title  string `json:"title"`
	Number int    `json:"number"`
	State  string `json:"state"`
}

type RepoPullRe struct {
	Name         string        `json:"name"`
	Private      bool          `json:"private"`
	Owner        string        `json:"owner"`
	PullRequests []PullRequest `json:"pull_requests"`
}
