package Tests

import (
	. "GoCalculator2.0-Distributed/Internal/Orchestrator"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddExpressionHandler(t *testing.T) {
	reqBody := []byte(`{"expression": "3 + 5 * (2 - 8)"}`)
	req, err := http.NewRequest("POST", "/api/v1/expressions", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(AddExpressionHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusCreated, status)
	}

	var response map[string]int
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Ошибка разбора JSON-ответа: %v", err)
	}

	if _, exists := response["id"]; !exists {
		t.Errorf("В ответе отсутствует ID выражения")
	}
}

func TestGetExpressionsHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/expressions", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetExpressionsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, status)
	}
}

func TestGetExpressionByIDHandler_NotFound(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/expressions/9999", nil) // Несуществующий ID
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetExpressionByIDHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusNotFound, status)
	}
}
