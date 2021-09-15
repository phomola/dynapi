package main

import "sync"

type book struct {
	Id     string `json:"id"`
	Author string `json:"author"`
	Title  string `json:"title"`
	Year   int    `json:"year"`
}

type BookService struct {
	booksMap map[string]*book
	mtx      sync.RWMutex
}

func NewBookService() *BookService {
	return &BookService{booksMap: make(map[string]*book)}
}
