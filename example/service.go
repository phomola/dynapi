package main

// StatusResponse is a status response.
type StatusResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
