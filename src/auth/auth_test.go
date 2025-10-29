package auth_test

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httptest"
	"testing"
	"vsC1Y2025V01/src/auth"
	"vsC1Y2025V01/src/db"
)

func TestRegisterAndLogin(t *testing.T) {
	logger := logrus.NewEntry(logrus.StandardLogger())

	// Initialize test DB (in-memory SQLite, Postgres test, etc.)
	db.InitDB(logger) // 💥 THIS IS MISSING

	// Setup test user payload
	payload := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	body, _ := json.Marshal(payload)

	// Test registration
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler := auth.RegisterHandler(logger)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	// Test login
	req = httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	auth.LoginHandler(logger).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
