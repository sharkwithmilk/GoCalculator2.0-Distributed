package main

import (
	"os"
	"strconv"
	"sync"
	"agent"
)

func main() {
	numAgents, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if numAgents <= 0 {
		numAgents = 2 // По умолчанию 2 агента
	}

	baseURL := "http://localhost"
	wg := &sync.WaitGroup{}

	for i := 1; i <= numAgents; i++ {
		agent := &Agent{ID: i, BaseURL: baseURL, WG: wg}
		wg.Add(1)
		go agent.Run()
	}

	wg.Wait()
}