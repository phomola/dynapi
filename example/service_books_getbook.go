package main

import "github.com/phomola/dynapi"

type getBookParams struct {
	Id string
}

type getBookResponse struct {
	statusResponse
	Book *book `json:"book,omitempty"`
}

func (s *BookService) GetBook(params *getBookParams, arg *dynapi.None) (*getBookResponse, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	if book, ok := s.booksMap[params.Id]; ok {
		return &getBookResponse{statusResponse: statusResponse{Status: "success"}, Book: book}, nil
	} else {
		return &getBookResponse{statusResponse: statusResponse{Status: "error", Error: "no book with this ID exists"}}, nil
	}
}
