package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
)

type Person struct {
	Name    string `json:"name"`
	Surname string `json:"surname"`
}

type PeopleStore struct {
	mu     sync.Mutex
	people []Person
}

func (s *PeopleStore) AddPerson(name, surname string) (Person, bool) {
	name = strings.TrimSpace(name)
	surname = strings.TrimSpace(surname)
	if name == "" || surname == "" {
		return Person{}, false
	}

	person := Person{
		Name:    name,
		Surname: surname,
	}

	s.mu.Lock()
	s.people = append(s.people, person)
	s.mu.Unlock()

	return person, true
}

func main() {
	store := &PeopleStore{
		people: make([]Person, 0),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/people", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req Person
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}

		person, ok := store.AddPerson(req.Name, req.Surname)
		if !ok {
			http.Error(w, "name and surname are required", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(person); err != nil {
			log.Printf("encode response error: %v", err)
		}
	})

	log.Println("server started on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
