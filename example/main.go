package main

import (
	"net/http"
	"sync"

	"github.com/phomola/dynapi"
)

type book struct {
	Id     string `json:"id"`
	Author string `json:"author"`
	Title  string `json:"title"`
	Year   int    `json:"year"`
}

type getBooksResponse struct {
	Books []*book `json:"books"`
}

type getBookParams struct {
	Id   string
	Test complex64
}

type getBookResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Book   *book  `json:"book,omitempty"`
}

type postBookResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type BookService struct {
	booksMap map[string]*book
	mtx      sync.RWMutex
}

func NewBookService() *BookService {
	return &BookService{booksMap: make(map[string]*book)}
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

func (s *BookService) GetBook(params *getBookParams, arg *dynapi.None) (*getBookResponse, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	if book, ok := s.booksMap[params.Id]; ok {
		return &getBookResponse{Status: "success", Book: book}, nil
	} else {
		return &getBookResponse{Status: "error", Error: "no book with this ID exists"}, nil
	}
}

func (s *BookService) PostBook(params *dynapi.None, book *book) (*postBookResponse, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, ok := s.booksMap[book.Id]; ok {
		return &postBookResponse{Status: "error", Error: "a book with this ID already exists"}, nil
	}
	s.booksMap[book.Id] = book
	return &postBookResponse{Status: "success"}, nil
}

func main() {
	mux := dynapi.New()
	s := NewBookService()
	if err := mux.HandleService("/api", s); err != nil {
		panic(err)
	}
	if err := http.ListenAndServe(":8080", mux.Handler()); err != nil {
		panic(err)
	}
}
