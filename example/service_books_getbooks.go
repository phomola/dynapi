package main

import "github.com/phomola/dynapi"

type getBooksResponse struct {
	Books []*book `json:"books"`
}

func (s *BookService) GetBooks(params *dynapi.None, arg *dynapi.None) (*getBooksResponse, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	books := make([]*book, 0, len(s.booksMap))
	for _, b := range s.booksMap {
		books = append(books, b)
	}
	return &getBooksResponse{Books: books}, nil
}
