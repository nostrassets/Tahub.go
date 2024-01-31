package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// Mock LndhubService for testing
type MockLndhubService struct{}

func (svc *MockLndhubService) CreateUser(ctx echo.Context, pubkey string) (*User, error) {
	// Mock implementation for CreateUser
	return &User{ID: "123", Pubkey: pubkey}, nil
}

// Mock responses for testing
var mockResponses = map[string]string{
	"TAHUB_CREATE_USER": `{
		"ID": "123",
		"Pubkey": "mock_pubkey",
		"CreatedAt": 1234567890,
		"Kind": 1,
		"Content": "TAHUB_CREATE_USER",
		"Sig": "mock_signature"
	}`,
	"TAHUB_RECEIVE_ADDRESS_FOR_ASSET": `{
		"ID": "456",
		"Pubkey": "mock_pubkey",
		"CreatedAt": 1234567890,
		"Kind": 1,
		"Content": "TAHUB_RECEIVE_ADDRESS_FOR_ASSET:ASSET_ID:AMT",
		"Sig": "mock_signature"
	}`,
	"TAHUB_SEND_ASSET": `{
		"ID": "789",
		"Pubkey": "mock_pubkey",
		"CreatedAt": 1234567890,
		"Kind": 1,
		"Content": "TAHUB_SEND_ASSET:ADDR:FEE",
		"Sig": "mock_signature"
	}`,
	"TAHUB_GET_BALANCES": `{
		"ID": "101112",
		"Pubkey": "mock_pubkey",
		"CreatedAt": 1234567890,
		"Kind": 1,
		"Content": "TAHUB_GET_BALANCES",
		"Sig": "mock_signature"
	}`,
}

func TestValidateNostrEventPayload(t *testing.T) {
	// Create a new Echo instance
	e := echo.New()

	// Marshal the mock response for TAHUB_CREATE_USER
	mockResponse := mockResponses["TAHUB_CREATE_USER"]
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(mockResponse))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	// Create a new HTTP response recorder
	rec := httptest.NewRecorder()

	// Create a new Echo context
	c := e.NewContext(req, rec)

	// Create a mock LndhubService
	mockService := &MockLndhubService{}

	// Create an instance of LndhubService with the mock service
	service := &LndhubService{svc: mockService}

	// Set up the middleware
	middleware := service.ValidateNostrEventPayload()

	// Define a handler function for testing
	handler := func(c echo.Context) error {
		// Your actual handler logic goes here
		return c.JSON(http.StatusOK, "OK")
	}

	// Call the middleware and pass the handler
	err := middleware(handler)(c)

	// Assertions

	// Check if there's no error
	assert.NoError(t, err)

	// Check if the response code is OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check if the validated payload is stored in the context
	var validatedPayload EventRequestBody
	err = json.Unmarshal([]byte(mockResponse), &validatedPayload)
	assert.NoError(t, err)
	assert.Equal(t, validatedPayload, c.Get("validatedPayload").(EventRequestBody))
}
