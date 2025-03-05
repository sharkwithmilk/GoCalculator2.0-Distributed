package main

import (
	. "GoCalculator2.0-Distributed/Internal/Orchestrator"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/api/v1/calculate", AddExpressionHandler)
	http.HandleFunc("/api/v1/expressions", GetExpressionsHandler)
	http.HandleFunc("/api/v1/expressions/", GetExpressionByIDHandler)
	http.HandleFunc("/internal/task", TaskHandler) // Обрабатывает и GET, и POST

	go CreateTasks()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Оркестратор запущен на порту", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
