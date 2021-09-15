package main

type statusResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
