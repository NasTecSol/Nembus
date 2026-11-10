package repository

// Response is a standard API response structure
type Response struct {
	StatusCode int         `json:"statusCode"` // e.g., 200, 500
	Message    string      `json:"message"`    // descriptive message
	Data       interface{} `json:"data"`       // can hold any type of data
}
