package main

import "sync"

type book struct {
	ID     string `json:"id"`
	Author string `json:"author"`
	Title  string `json:"title"`
	Year   int    `json:"year"`
}

// BookService is a book service.
type BookService struct {
	booksMap map[string]*book
	mtx      sync.RWMutex
}

// NewBookService creates a new book service.
func NewBookService() *BookService {
	return &BookService{booksMap: make(map[string]*book)}
}
