package main

import (
	"fmt"
	"log"
	"net/http"
	"orchestrator"
)

func main() {
	http.HandleFunc("/api/v1/calculate", addExpressionHandler)
	http.HandleFunc("/api/v1/expressions", getExpressionsHandler)
	http.HandleFunc("/internal/task", getTaskHandler)
	http.HandleFunc("/internal/task", submitTaskHandler)

	go createTasks()

	fmt.Println("Оркестратор запущен на порту 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}