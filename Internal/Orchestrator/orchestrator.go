package Orchestrator

import (
	t "GoCalculator2.0-Distributed/Pkg"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Expression struct {
	ID     int     `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result,omitempty"`
	Tokens []t.Token
}

var (
	expressions = make(map[int]*Expression)
	tasks       = make(chan t.Task, 100)
	mutex       sync.Mutex
	exprID      = 1
	taskID      = 1
)

func AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	mutex.Lock()
	id := exprID
	exprID++
	mutex.Unlock()

	tokens := t.Tokenize(req.Expression)
	expressions[id] = &Expression{
		ID:     id,
		Status: "pending",
		Tokens: tokens,
	}

	fmt.Printf("[Оркестратор] Добавлено выражение ID %d: %s\n", id, req.Expression)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

func GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	var list []struct {
		ID     int     `json:"id"`
		Status string  `json:"status"`
		Result float64 `json:"result,omitempty"`
	}
	for _, expr := range expressions {
		list = append(list, struct {
			ID     int     `json:"id"`
			Status string  `json:"status"`
			Result float64 `json:"result,omitempty"`
		}{
			ID:     expr.ID,
			Status: expr.Status,
			Result: expr.Result,
		})
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": list})
}

func GetExpressionByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/v1/expressions/"):] // Получаем ID из пути
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	expr, exists := expressions[id]
	mutex.Unlock()

	if !exists {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	// Формируем JSON без токенов
	response := struct {
		ID     int     `json:"id"`
		Status string  `json:"status"`
		Result float64 `json:"result,omitempty"`
	}{
		ID:     expr.ID,
		Status: expr.Status,
		Result: expr.Result,
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"expression": response})
}

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		select {
		case task := <-tasks:
			json.NewEncoder(w).Encode(map[string]t.Task{"task": task})
		default:
			http.Error(w, "No task available", http.StatusNotFound)
		}
	} else if r.Method == http.MethodPost {
		var res struct {
			ID     int     `json:"id"`
			Result float64 `json:"result"`
			Error  string  `json:"error,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
			http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
			return
		}

		mutex.Lock()
		defer mutex.Unlock()

		for _, expr := range expressions {
			if expr.ID == res.ID {
				if res.Error != "" {
					expr.Status = "error"
					fmt.Printf("[Оркестратор] Ошибка при вычислении выражения ID %d: %s\n", res.ID, res.Error)
				} else {
					expr.Status = "completed"
					expr.Result = res.Result
					fmt.Printf("[Оркестратор] Выражение ID %d вычислено: %f\n", res.ID, res.Result)
				}
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		http.Error(w, "Task not found", http.StatusNotFound)
	}
}

func CreateTasks() {
	for {
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
		mutex.Lock()
		for id, expr := range expressions {
			if expr.Status == "pending" {
				root := (&t.Parser{Tokens: expr.Tokens}).ParseExpression()
				task := t.Task{
					ID:            taskID,
					Arg1:          root.Left.Value,
					Arg2:          root.Right.Value,
					Operation:     root.Operator,
					OperationTime: rand.Intn(500),
				}
				taskID++
				tasks <- task
				expr.Status = "in_progress"
				fmt.Printf("[Оркестратор] Создана задача ID %d для выражения %d\n", task.ID, id)
			}
		}
		mutex.Unlock()
	}
}
