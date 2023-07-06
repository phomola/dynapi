package main

import "github.com/phomola/dynapi"

// GetBooksResponse is a get books response.
type GetBooksResponse struct {
	Books []*book `json:"books"`
}

// GetBooks retrieves all books.
func (s *BookService) GetBooks(ctx *dynapi.HandlerContext, params *dynapi.None, arg *dynapi.None) (*GetBooksResponse, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	books := make([]*book, 0, len(s.booksMap))
	for _, b := range s.booksMap {
		books = append(books, b)
	}
	return &GetBooksResponse{Books: books}, nil
}
