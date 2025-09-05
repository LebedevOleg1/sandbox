package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Task statuses
const (
	StatusInProgress = "in_progress"
	StatusReady      = "ready"
)

// Task model
type Task struct {
	ID     string `json:"task_id"`
	Status string `json:"status"`
	Result string `json:"result,omitempty"`
}

// In-memory storage
type TaskStore struct {
	sync.RWMutex
	tasks map[string]*Task
}

func NewTaskStore() *TaskStore {
	return &TaskStore{tasks: make(map[string]*Task)}
}

func (s *TaskStore) CreateTask() *Task {
	s.Lock()
	defer s.Unlock()

	id := uuid.New().String()
	task := &Task{ID: id, Status: StatusInProgress}
	s.tasks[id] = task

	// имитация выполнения кода (через горутину)
	go func() {
		time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)
		s.Lock()
		task.Status = StatusReady
		task.Result = "Fake result: Hello from sandbox!"
		s.Unlock()
	}()

	return task
}

func (s *TaskStore) GetTask(id string) (*Task, bool) {
	s.RLock()
	defer s.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

func main() {
	store := NewTaskStore()

	http.HandleFunc("/task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		task := store.CreateTask()
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"task_id": task.ID})
	})

	http.HandleFunc("/status/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/status/"):]
		task, ok := store.GetTask(id)
		if !ok {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": task.Status})
	})

	http.HandleFunc("/result/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/result/"):]
		task, ok := store.GetTask(id)
		if !ok {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"result": task.Result})
	})

	log.Println("Server is running on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
