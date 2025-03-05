package Tests

import (
	a "GoCalculator2.0-Distributed/Internal/Agent"
	te "GoCalculator2.0-Distributed/Pkg"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestAgentProcessTask_Addition(t *testing.T) {
	agent := a.Agent{ID: 1, BaseURL: "http://localhost:8080", WG: &sync.WaitGroup{}}
	task := &te.Task{ID: 1, Arg1: 5, Arg2: 3, Operation: "+"}

	result, err := agent.ProcessTask(task)
	if err != nil {
		t.Fatalf("Ошибка при обработке задачи: %v", err)
	}

	expected := 8.0
	if result != expected {
		t.Errorf("Ожидался результат %f, получено %f", expected, result)
	}
}

func TestAgentProcessTask_DivisionByZero(t *testing.T) {
	agent := a.Agent{ID: 2, BaseURL: "http://localhost:8080", WG: &sync.WaitGroup{}}
	task := &te.Task{ID: 2, Arg1: 10, Arg2: 0, Operation: "/"}

	_, err := agent.ProcessTask(task)
	if err == nil {
		t.Fatalf("Ожидалась ошибка деления на ноль, но её не было")
	}
}

func TestFetchTask_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "No task available", http.StatusNotFound)
	}))
	defer server.Close()

	agent := a.Agent{ID: 3, BaseURL: server.URL, WG: &sync.WaitGroup{}}
	task, err := agent.FetchTask()

	if err == nil {
		t.Errorf("Ожидалась ошибка при отсутствии задачи, но её нет")
	}

	if task != nil {
		t.Errorf("Ожидалось nil, но получено: %+v", task)
	}
}
