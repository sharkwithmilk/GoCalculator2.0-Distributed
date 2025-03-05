package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Task struct {
	ID            int     `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

type Agent struct {
	ID      int
	BaseURL string
	WG      *sync.WaitGroup
}

func (a *Agent) fetchTask() (*Task, error) {
	resp, err := http.Get(a.BaseURL + "/internal/task")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("no tasks available")
	}

	var response struct {
		Task Task `json:"task"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response.Task, nil
}

func (a *Agent) processTask(task *Task) float64 {
	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2
	case "-":
		return task.Arg1 - task.Arg2
	case "*":
		return task.Arg1 * task.Arg2
	case "/":
		if task.Arg2 != 0 {
			return task.Arg1 / task.Arg2
		}
	}
	return 0
}

func (a *Agent) submitResult(taskID int, result float64) {
	data, _ := json.Marshal(map[string]interface{}{
		"id":     taskID,
		"result": result,
	})
	_, err := http.Post(a.BaseURL+"/internal/task", "application/json", 
		bytes.NewBuffer(data))
	if err != nil {
		log.Printf("[Агент %d] Ошибка отправки результата: %v", a.ID, err)
	} else {
		log.Printf("[Агент %d] Результат отправлен: %d -> %f", a.ID, taskID, result)
	}
}

func (a *Agent) Run() {
	defer a.WG.Done()
	for {
		task, err := a.fetchTask()
		if err != nil {
			log.Printf("[Агент %d] Нет задач, жду...", a.ID)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("[Агент %d] Получена задача: %v", a.ID, task)
		time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)
		result := a.processTask(task)
		a.submitResult(task.ID, result)
	}
}