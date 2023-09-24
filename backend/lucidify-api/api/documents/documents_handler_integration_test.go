package documents

import (
	"bytes"
	"encoding/json"
	"lucidify-api/modules/clerkclient"
	"lucidify-api/modules/config"
	"lucidify-api/modules/store"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDocumentsUploadHandlerIntegration(t *testing.T) {
	testconfig := config.NewServerConfig()
	PostgresqlURL := testconfig.PostgresqlURL
	db, err := store.NewStore(PostgresqlURL)
	// Setup the real environment
	clerkInstance, err := clerkclient.NewClerkClient(testconfig.ClerkSecretKey)

	if err != nil {
		t.Fatalf("Failed to create Clerk client: %v", err)
	}
	cfg := &config.ServerConfig{}

	// Create a test server
	mux := http.NewServeMux()
	SetupRoutes(cfg, mux, db, clerkInstance)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Obtain a JWT token from Clerk
	jwtToken := testconfig.TestJWTSessionToken

	// Send a POST request to the server with the JWT token
	document := map[string]string{
		"document_name": "Test Document",
		"content":       "Test Content",
	}
	body, _ := json.Marshal(document)
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/documents/upload", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var responseBody map[string]string
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&responseBody)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	expectedResponse := "PLACEHOLDER RESPONSE"
	if responseBody["response"] != expectedResponse {
		t.Errorf("Expected response %s, got %s", expectedResponse, responseBody["response"])
	}

	// Optionally: Check the database to ensure the document was saved correctly
}
