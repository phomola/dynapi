package main

import "github.com/phomola/dynapi"

func (s *BookService) PostBook(ctx *dynapi.HandlerContext, params *dynapi.None, book *book) (*statusResponse, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, ok := s.booksMap[book.Id]; ok {
		return &statusResponse{Status: "error", Error: "a book with this ID already exists"}, nil
	}
	s.booksMap[book.Id] = book
	return &statusResponse{Status: "success"}, nil
}
