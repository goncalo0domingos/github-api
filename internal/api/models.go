package api

// repo variables
type Repository struct {
	Name string `json:"name" binding:"required"`
}

type RepoPrint struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
}
