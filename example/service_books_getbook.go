package main

import "github.com/phomola/dynapi"

type getBookParams struct {
	ID string
}

// GetBookResponse is a get book response.
type GetBookResponse struct {
	StatusResponse
	Book *book `json:"book,omitempty"`
}

// GetBook retrieves the book with the given ID.
func (s *BookService) GetBook(ctx *dynapi.HandlerContext, params *getBookParams, arg *dynapi.None) (*GetBookResponse, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	if book, ok := s.booksMap[params.ID]; ok {
		return &GetBookResponse{StatusResponse: StatusResponse{Status: "success"}, Book: book}, nil
	}
	return &GetBookResponse{StatusResponse: StatusResponse{Status: "error", Error: "no book with this ID exists"}}, nil
}
