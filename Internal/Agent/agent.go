package Agent

import (
	t "GoCalculator2.0-Distributed/Pkg"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Agent struct {
	ID      int
	BaseURL string
	WG      *sync.WaitGroup
}

func (a *Agent) FetchTask() (*t.Task, error) {
	resp, err := http.Get(a.BaseURL + "/internal/task")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("no tasks available")
	}

	var response struct {
		Task t.Task `json:"task"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response.Task, nil
}

func getOperationTime(operation string) int {
	var envVar string
	switch operation {
	case "+":
		envVar = "TIME_ADDITION_MS"
	case "-":
		envVar = "TIME_SUBTRACTION_MS"
	case "*":
		envVar = "TIME_MULTIPLICATION_MS"
	case "/":
		envVar = "TIME_DIVISION_MS"
	}

	timeMs, err := strconv.Atoi(os.Getenv(envVar))
	if err != nil {
		return 100 // Значение по умолчанию
	}
	return timeMs
}

func (a *Agent) ProcessTask(task *t.Task) (float64, error) {
	time.Sleep(time.Duration(getOperationTime(task.Operation)) * time.Millisecond)
	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2, nil
	case "-":
		return task.Arg1 - task.Arg2, nil
	case "*":
		return task.Arg1 * task.Arg2, nil
	case "/":
		if task.Arg2 != 0 {
			return task.Arg1 / task.Arg2, nil
		}
		return 0, fmt.Errorf("division by zero")
	}
	return 0, fmt.Errorf("unknown operation")
}

func (a *Agent) submitResult(taskID int, result float64, err error) {
	data := map[string]interface{}{
		"id":     taskID,
		"result": result,
	}
	if err != nil {
		data["error"] = err.Error()
	}

	jsonData, _ := json.Marshal(data)
	log.Printf("[Агент %d] Отправка результата: %s", a.ID, jsonData)

	_, postErr := http.Post(a.BaseURL+"/internal/task", "application/json", bytes.NewBuffer(jsonData))
	if postErr != nil {
		log.Printf("[Агент %d] Ошибка отправки результата: %v", a.ID, postErr)
	} else {
		log.Printf("[Агент %d] Результат отправлен: %d -> %f, Ошибка: %v", a.ID, taskID, result, err)
	}
}

func (a *Agent) Run() {
	defer a.WG.Done()
	for {
		task, err := a.FetchTask()
		if err != nil {
			log.Printf("[Агент %d] Нет задач, жду...", a.ID)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("[Агент %d] Получена задача: %v", a.ID, task)
		result, calcErr := a.ProcessTask(task)
		a.submitResult(task.ID, result, calcErr)
	}
}
