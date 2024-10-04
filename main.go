package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Book represents a book item with an ID, title, author, and price.
type Book struct {
	ID     int     `json:"id"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float64 `json:"price"`
}

// Global variables to store book items and synchronize access.
var (
	books  = make(map[int]Book)
	nextID = 1
	mu     sync.Mutex
)

func main() {
	// Setting up handlers for books and specific book actions.
	http.HandleFunc("/books", booksHandler)
	http.HandleFunc("/books/", bookHandler) // For specific book actions (get, update, delete)
	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", nil)
}

// booksHandler handles general book collection operations (GET, POST).
func booksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getBooks(w)
	case http.MethodPost:
		createBook(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// bookHandler handles operations on a specific book (GET, PUT, DELETE).
func bookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getBook(w, id)
	case http.MethodPut:
		updateBook(w, r, id)
	case http.MethodDelete:
		deleteBook(w, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getBooks retrieves the list of all books.
func getBooks(w http.ResponseWriter) {
	mu.Lock()
	defer mu.Unlock()

	bookList := make([]Book, 0, len(books))
	for _, book := range books {
		bookList = append(bookList, book)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookList)
}

// createBook creates a new book and adds it to the collection.
func createBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	mu.Lock()
	book.ID = nextID
	nextID++
	books[book.ID] = book
	mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

// getBook retrieves a specific book by its ID.
func getBook(w http.ResponseWriter, id int) {
	mu.Lock()
	defer mu.Unlock()

	book, found := books[id]
	if !found {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

// updateBook updates an existing book's details.
func updateBook(w http.ResponseWriter, r *http.Request, id int) {
	mu.Lock()
	defer mu.Unlock()

	book, found := books[id]
	if !found {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	books[id] = book
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

// deleteBook removes a book from the collection.
func deleteBook(w http.ResponseWriter, id int) {
	mu.Lock()
	defer mu.Unlock()

	if _, found := books[id]; !found {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	delete(books, id)
	w.WriteHeader(http.StatusNoContent)
}

// parseID extracts the ID from the URL path.
func parseID(path string) (int, error) {
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return 0, fmt.Errorf("invalid path")
	}

	id, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, fmt.Errorf("invalid book ID")
	}
	return id, nil
}
