package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/josestg/mux"
)

func main() {

	h := newBookHandler()

	m := mux.New()

	m.HandleFunc(http.MethodGet, "/ping", func(w http.ResponseWriter, r *http.Request) {
		responseJSON(w, http.StatusOK, map[string]string{"msg": "OK"})
	})

	m.HandleFunc(http.MethodGet, "/books", h.list)
	m.HandleFunc(http.MethodPost, "/books", h.create)
	m.HandleFunc(http.MethodGet, "/books/:id", h.detail)
	m.HandleFunc(http.MethodDelete, "/books/:id", h.delete)

	// add global middleware
	m.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// do middleware stuff here
			log.Println("doing middleware stuff here")

			// call the next handler, which can be the next middleware or the final handler
			next.ServeHTTP(w, r)
		})
	})

	log.Fatal(http.ListenAndServe("0.0.0.0:8080", m))
}

type book struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	DateCreated time.Time `json:"date_created"`
}

type bookHandler struct {
	repo *bookStorage
}

func newBookHandler() *bookHandler {
	return &bookHandler{
		repo: &bookStorage{
			books: make(map[int]book),
			mutex: new(sync.Mutex),
		},
	}
}

func (h *bookHandler) list(w http.ResponseWriter, r *http.Request) {
	books := h.repo.all()
	responseJSON(w, http.StatusOK, books)
}

func (h *bookHandler) create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		responseJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	id := len(h.repo.all()) + 1
	newBook := book{ID: id, Title: input.Title, DateCreated: time.Now()}
	h.repo.add(newBook)

	responseJSON(w, http.StatusCreated, newBook)
}

func (h *bookHandler) detail(w http.ResponseWriter, r *http.Request) {
	vars := mux.GetVars(r.Context())
	id, _ := strconv.Atoi(vars.Get("id"))

	b, ok := h.repo.get(id)
	if !ok {
		responseJSON(w, http.StatusNotFound, map[string]string{"error": "Not Found"})
		return
	}

	responseJSON(w, http.StatusOK, b)
}

func (h *bookHandler) delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.GetVars(r.Context())
	id, _ := strconv.Atoi(vars.Get("id"))

	b, ok := h.repo.get(id)
	if !ok {
		responseJSON(w, http.StatusNotFound, map[string]string{"error": "Not Found"})
		return
	}

	h.repo.delete(id)
	responseJSON(w, http.StatusOK, b)
}

func responseJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

type bookStorage struct {
	books map[int]book
	mutex *sync.Mutex
}

func (bs *bookStorage) add(b book) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	bs.books[b.ID] = b
}

func (bs *bookStorage) get(id int) (book, bool) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	b, ok := bs.books[id]
	return b, ok
}

func (bs *bookStorage) delete(id int) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	delete(bs.books, id)
}

func (bs *bookStorage) all() []book {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	books := make([]book, 0, len(bs.books))

	for _, v := range bs.books {
		books = append(books, v)
	}

	return books
}
