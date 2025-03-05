package orchestrator

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"tokenization"
)

type Expression struct {
	ID       int     `json:"id"`
	Status   string  `json:"status"`
	Result   float64 `json:"result,omitempty"`
	Tokens   []tokenization.Token
}

type Task struct {
	ID            int     `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

var (
	expressions = make(map[int]*Expression)
	tasks       = make(chan Task, 100)
	mutex       sync.Mutex
	exprID      = 1
	taskID      = 1
)

func addExpressionHandler(w http.ResponseWriter, r *http.Request) {
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

	tokens := tokenization.Tokenize(req.Expression)
	expressions[id] = &Expression{
		ID:     id,
		Status: "pending",
		Tokens: tokens,
	}

	fmt.Printf("[Оркестратор] Добавлено выражение ID %d: %s\n", id, req.Expression)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

func getExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	var list []Expression
	for _, expr := range expressions {
		list = append(list, *expr)
	}
	json.NewEncoder(w).Encode(map[string][]Expression{"expressions": list})
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case task := <-tasks:
		json.NewEncoder(w).Encode(map[string]Task{"task": task})
	default:
		http.Error(w, "No task available", http.StatusNotFound)
	}
}

func submitTaskHandler(w http.ResponseWriter, r *http.Request) {
	var res struct {
		ID     int     `json:"id"`
		Result float64 `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	for _, expr := range expressions {
		if expr.ID == res.ID {
			expr.Status = "completed"
			expr.Result = res.Result
			fmt.Printf("[Оркестратор] Выражение ID %d вычислено: %f\n", res.ID, res.Result)
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	http.Error(w, "Task not found", http.StatusNotFound)
}

func createTasks() {
	for {
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
		mutex.Lock()
		for id, expr := range expressions {
			if expr.Status == "pending" {
				root := tokenization.Parser{Tokens: expr.Tokens}.ParseExpression()
				task := Task{
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