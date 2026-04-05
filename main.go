package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
)

type Person struct {
	Name           string `json:"name"`
	Surname        string `json:"surname"`
	StudyDirection string `json:"studyDirection"`
	VisitedEvent   bool   `json:"visitedEvent"`
}

type PeopleStore struct {
	mu     sync.Mutex
	people []Person
}

var (
	errPersonNotFound      = errors.New("person not found")
	errPersonAlreadyPassed = errors.New("such person already passed")
)

func (s *PeopleStore) AddPerson(name, surname, studyDirection string, visitedEvent bool) (Person, bool) {
	name = strings.TrimSpace(name)
	surname = strings.TrimSpace(surname)
	studyDirection = strings.TrimSpace(studyDirection)
	if name == "" || surname == "" || studyDirection == "" {
		return Person{}, false
	}

	person := Person{
		Name:           name,
		Surname:        surname,
		StudyDirection: studyDirection,
		VisitedEvent:   visitedEvent,
	}

	s.mu.Lock()
	s.people = append(s.people, person)
	s.mu.Unlock()

	return person, true
}

func (s *PeopleStore) MarkPersonAsVisited(name, surname string) (Person, error) {
	name = strings.TrimSpace(name)
	surname = strings.TrimSpace(surname)
	if name == "" || surname == "" {
		return Person{}, errPersonNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.people {
		if s.people[i].Name != name || s.people[i].Surname != surname {
			continue
		}
		if s.people[i].VisitedEvent {
			return Person{}, errPersonAlreadyPassed
		}

		s.people[i].VisitedEvent = true
		return s.people[i], nil
	}

	return Person{}, errPersonNotFound
}

type checkInRequest struct {
	Name    string `json:"name"`
	Surname string `json:"surname"`
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

		person, ok := store.AddPerson(req.Name, req.Surname, req.StudyDirection, req.VisitedEvent)
		if !ok {
			http.Error(w, "name, surname and studyDirection are required", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(person); err != nil {
			log.Printf("encode response error: %v", err)
		}
	})
	mux.HandleFunc("/people/check-in", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req checkInRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}

		person, err := store.MarkPersonAsVisited(req.Name, req.Surname)
		if err != nil {
			switch err {
			case errPersonAlreadyPassed:
				http.Error(w, "such person already passed", http.StatusConflict)
			case errPersonNotFound:
				http.Error(w, "person not found", http.StatusNotFound)
			default:
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(person); err != nil {
			log.Printf("encode response error: %v", err)
		}
	})

	log.Println("server started on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
