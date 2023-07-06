package main

import "github.com/phomola/dynapi"

// PostBook creates a book.
func (s *BookService) PostBook(ctx *dynapi.HandlerContext, params *dynapi.None, book *book) (*StatusResponse, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, ok := s.booksMap[book.ID]; ok {
		return &StatusResponse{Status: "error", Error: "a book with this ID already exists"}, nil
	}
	s.booksMap[book.ID] = book
	return &StatusResponse{Status: "success"}, nil
}
