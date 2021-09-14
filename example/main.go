package main

import (
	"net/http"
	"sync"

	"github.com/phomola/dynapi"
)

var (
	booksMap = make(map[string]*book)
	mtx      sync.RWMutex
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
	Id string
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

func getBooks(params *dynapi.None, arg *dynapi.None) (*getBooksResponse, error) {
	mtx.RLock()
	defer mtx.RUnlock()
	books := make([]*book, 0, len(booksMap))
	for _, b := range booksMap {
		books = append(books, b)
	}
	return &getBooksResponse{Books: books}, nil
}

func getBook(params *getBookParams, arg *dynapi.None) (*getBookResponse, error) {
	mtx.RLock()
	defer mtx.RUnlock()
	if book, ok := booksMap[params.Id]; ok {
		return &getBookResponse{Status: "success", Book: book}, nil
	} else {
		return &getBookResponse{Status: "error", Error: "no book with this ID exists"}, nil
	}
}

func postBook(params *dynapi.None, book *book) (*postBookResponse, error) {
	mtx.Lock()
	defer mtx.Unlock()
	if _, ok := booksMap[book.Id]; ok {
		return &postBookResponse{Status: "error", Error: "a book with this ID already exists"}, nil
	}
	booksMap[book.Id] = book
	return &postBookResponse{Status: "success"}, nil
}

func main() {
	mux := dynapi.New()
	err := mux.Handle(getBooks)
	if err != nil {
		panic(err)
	}
	err = mux.Handle(getBook)
	if err != nil {
		panic(err)
	}
	err = mux.Handle(postBook)
	if err != nil {
		panic(err)
	}
	err = http.ListenAndServe(":8080", mux.Handler())
	if err != nil {
		panic(err)
	}
}
