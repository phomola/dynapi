package main

import (
	"net/http"

	"github.com/phomola/dynapi"
)

func main() {
	mux := dynapi.New()
	s := NewBookService()
	if err := mux.HandleService("/api", s); err != nil {
		panic(err)
	}
	mux.FinishSetup()
	if err := http.ListenAndServe(":8080", mux.Handler()); err != nil {
		panic(err)
	}
}
