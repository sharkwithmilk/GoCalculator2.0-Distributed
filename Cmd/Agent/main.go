package main

import (
	. "GoCalculator2.0-Distributed/Internal/Agent"
	"os"
	"strconv"
	"sync"
)

func main() {
	numAgents, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if numAgents <= 0 {
		numAgents = 2 // По умолчанию 2 агента
	}

	baseURL := "http://localhost:8080"
	wg := &sync.WaitGroup{}

	for i := 1; i <= numAgents; i++ {
		agent := &Agent{ID: i, BaseURL: baseURL, WG: wg}
		wg.Add(1)
		go agent.Run()
	}

	wg.Wait()
}
