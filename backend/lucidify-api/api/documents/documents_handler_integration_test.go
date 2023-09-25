package documents

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"lucidify-api/modules/clerkclient"
	"lucidify-api/modules/config"
	"lucidify-api/modules/store"
	"net/http"
	"net/http/httptest"
	"testing"
)

func createTestUserInDb() {
	testconfig := config.NewServerConfig()
	PostgresqlURL := testconfig.PostgresqlURL
	db, err := store.NewStore(PostgresqlURL)

	// the user id registered by the jwt token must exist in the local database
	user := store.User{
		UserID:           testconfig.TestUserID,
		ExternalID:       "TestCreateUserInUsersTableExternalID",
		Username:         "TestCreateUserInUsersTableUsername",
		PasswordEnabled:  true,
		Email:            "TestCreateUserInUsersTable@example.com",
		FirstName:        "TestCreateUserInUsersTableCreateTest",
		LastName:         "TestCreateUserInUsersTableUser",
		ImageURL:         "https://TestCreateUserInUsersTable.com/image.jpg",
		ProfileImageURL:  "https://TestCreateUserInUsersTable.com/profile.jpg",
		TwoFactorEnabled: false,
		CreatedAt:        1654012591514,
		UpdatedAt:        1654012591514,
	}

	db.DeleteUserInUsersTable(user.UserID)
	err = db.CheckUserDeletedInUsersTable(user.UserID, 3)
	if err != nil {
		log.Fatalf("Failed to delete user: %v", err)
	}
	err = db.CreateUserInUsersTable(user)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	// Check if the user exists
	err = db.CheckIfUserInUsersTable(user.UserID, 3)
	if err != nil {
		log.Fatalf("User not found after creation: %v", err)
	}
}

func TestDocumentsUploadHandlerIntegration(t *testing.T) {
	testconfig := config.NewServerConfig()
	PostgresqlURL := testconfig.PostgresqlURL
	db, err := store.NewStore(PostgresqlURL)
	// Setup the real environment
	clerkInstance, err := clerkclient.NewClerkClient(testconfig.ClerkSecretKey)
	createTestUserInDb()

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

	documentFromDb, err := db.GetDocument(testconfig.TestUserID, "Test Document")
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}

	documentFromDbContent := documentFromDb.Content
	if documentFromDbContent != "Test Content" {
		t.Fatalf("Expected document content %s, got %s", "Test Content", documentFromDbContent)
	}

	// Cleanup the database
	t.Cleanup(func() {
		testconfig := config.NewServerConfig()
		UserID := testconfig.TestUserID
		db.DeleteUserInUsersTable(UserID)
		db.DeleteDocument(UserID, "Test Document")
	})
}

func TestDocumentsUploadHandlerUnauthorizedIntegration(t *testing.T) {
	testconfig := config.NewServerConfig()
	PostgresqlURL := testconfig.PostgresqlURL
	db, err := store.NewStore(PostgresqlURL)
	// Setup the real environment
	clerkInstance, err := clerkclient.NewClerkClient(testconfig.ClerkSecretKey)
	createTestUserInDb()

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
	jwtToken := testconfig.TestJWTSessionToken + "invalid"

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
	if resp.StatusCode == http.StatusOK {
		t.Errorf("Expected status code not OK, got %d", resp.StatusCode)
	}

	// Cleanup the database
	t.Cleanup(func() {
		testconfig := config.NewServerConfig()
		UserID := testconfig.TestUserID
		db.DeleteUserInUsersTable(UserID)
		db.DeleteDocument(UserID, "Test Document")
	})
}

func TestDocumentsGetDocumentHandlerIntegration(t *testing.T) {
	testconfig := config.NewServerConfig()
	PostgresqlURL := testconfig.PostgresqlURL
	db, err := store.NewStore(PostgresqlURL)

	clerkInstance, err := clerkclient.NewClerkClient(testconfig.ClerkSecretKey)
	if err != nil {
		t.Fatalf("Failed to create Clerk client: %v", err)
	}

	createTestUserInDb()
	cfg := &config.ServerConfig{}

	// Create a test server
	mux := http.NewServeMux()
	SetupRoutes(cfg, mux, db, clerkInstance)
	server := httptest.NewServer(mux)
	defer server.Close()

	jwtToken := testconfig.TestJWTSessionToken

	document := map[string]string{
		"document_name": "Test Document",
		"content":       "Test Content",
	}

	db.UploadDocument(testconfig.TestUserID, "Test Document", "Test Content")

	body, _ := json.Marshal(document)
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/documents/getdocument", bytes.NewBuffer(body))
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

	// Please implement the rest of this integration test to check it returns the correct document
	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Unmarshal the response body into a Document object
	var respDocument store.Document
	err = json.Unmarshal(respBody, &respDocument)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	// Check if the returned document is correct
	if respDocument.DocumentName != document["document_name"] || respDocument.Content != document["content"] {
		t.Errorf("Returned document does not match the expected document")
	}

	// Cleanup the database
	t.Cleanup(func() {
		testconfig := config.NewServerConfig()
		UserID := testconfig.TestUserID
		db.DeleteUserInUsersTable(UserID)
		db.DeleteDocument(UserID, "Test Document")
	})
}

func TestDocumentsGetDocumentHandlerUnauthorizedIntegration(t *testing.T) {
	testconfig := config.NewServerConfig()
	PostgresqlURL := testconfig.PostgresqlURL
	db, err := store.NewStore(PostgresqlURL)

	clerkInstance, err := clerkclient.NewClerkClient(testconfig.ClerkSecretKey)

	createTestUserInDb()

	if err != nil {
		t.Fatalf("Failed to create Clerk client: %v", err)
	}
	cfg := &config.ServerConfig{}

	// Create a test server
	mux := http.NewServeMux()
	SetupRoutes(cfg, mux, db, clerkInstance)
	server := httptest.NewServer(mux)
	defer server.Close()

	jwtToken := testconfig.TestJWTSessionToken + "invalid"

	document := map[string]string{
		"document_name": "Test Document",
		"content":       "Test Content",
	}

	db.UploadDocument(testconfig.TestUserID, "Test Document", "Test Content")

	body, _ := json.Marshal(document)
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/documents/getdocument", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status code Bad Request, 400. Got: %v", resp.StatusCode)
	}

	// Cleanup the database
	t.Cleanup(func() {
		testconfig := config.NewServerConfig()
		UserID := testconfig.TestUserID
		db.DeleteUserInUsersTable(UserID)
		db.DeleteDocument(UserID, "Test Document")
	})
}

func TestDocumentsGetAllDocumentsHandlerIntegration(t *testing.T) {
	testconfig := config.NewServerConfig()
	PostgresqlURL := testconfig.PostgresqlURL
	db, err := store.NewStore(PostgresqlURL)

	clerkInstance, err := clerkclient.NewClerkClient(testconfig.ClerkSecretKey)

	createTestUserInDb()

	if err != nil {
		t.Fatalf("Failed to create Clerk client: %v", err)
	}
	cfg := &config.ServerConfig{}

	// Create a test server
	mux := http.NewServeMux()
	SetupRoutes(cfg, mux, db, clerkInstance)
	server := httptest.NewServer(mux)
	defer server.Close()

	jwtToken := testconfig.TestJWTSessionToken

	// document := map[string]string{
	// 	"document_name": "Test Document",
	// 	"content":       "Test Content",
	// }

	db.UploadDocument(testconfig.TestUserID, "Test Document", "Test Content")
	db.UploadDocument(testconfig.TestUserID, "Test Document 2", "Test Content 2")
	db.UploadDocument(testconfig.TestUserID, "Test Document 3", "Test Content 3")

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/documents/getalldocuments", nil)
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

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Unmarshal the response body into a slice of Document objects
	var respDocuments []store.Document
	err = json.Unmarshal(respBody, &respDocuments)
	if err != nil {
		t.Errorf("Failed to unmarshal response body: %v", err)
	}

	// Check if the returned documents are correct
	if len(respDocuments) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(respDocuments))
	}

	expectedDocs := []string{"Test Document", "Test Document 2", "Test Document 3"}
	for i, doc := range respDocuments {
		if doc.DocumentName != expectedDocs[i] {
			t.Errorf("Expected document name %s, got %s", expectedDocs[i], doc.DocumentName)
		}
	}

	// Cleanup the database
	t.Cleanup(func() {
		testconfig := config.NewServerConfig()
		UserID := testconfig.TestUserID
		db.DeleteUserInUsersTable(UserID)
		db.DeleteDocument(UserID, "Test Document")
		db.DeleteDocument(UserID, "Test Document 2")
		db.DeleteDocument(UserID, "Test Document 3")
	})
}

// func TestDocumentsGetAllDocumentsHandlerUnauthenticatedIntegration(t *testing.T) {
// 	testconfig := config.NewServerConfig()
// 	PostgresqlURL := testconfig.PostgresqlURL
// 	db, err := store.NewStore(PostgresqlURL)
//
// 	clerkInstance, err := clerkclient.NewClerkClient(testconfig.ClerkSecretKey)
//
// 	db.CheckUserDeletedInUsersTable(testconfig.TestUserID, 3)
// 	createTestUserInDb()
//
// 	if err != nil {
// 		t.Fatalf("Failed to create Clerk client: %v", err)
// 	}
// 	cfg := &config.ServerConfig{}
//
// 	// Create a test server
// 	mux := http.NewServeMux()
// 	SetupRoutes(cfg, mux, db, clerkInstance)
// 	server := httptest.NewServer(mux)
// 	defer server.Close()
//
// 	jwtToken := testconfig.TestJWTSessionToken + "invalid"
//
// 	// document := map[string]string{
// 	// 	"document_name": "Test Document",
// 	// 	"content":       "Test Content",
// 	// }
//
// 	db.UploadDocument(testconfig.TestUserID, "Test Document", "Test Content")
// 	db.UploadDocument(testconfig.TestUserID, "Test Document 2", "Test Content 2")
// 	db.UploadDocument(testconfig.TestUserID, "Test Document 3", "Test Content 3")
//
// 	req, _ := http.NewRequest(http.MethodGet, server.URL+"/documents/getalldocuments", nil)
// 	req.Header.Set("Authorization", "Bearer "+jwtToken)
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		t.Fatalf("Failed to send request: %v", err)
// 	}
// 	defer resp.Body.Close()
//
// 	// Check the response
// 	if resp.StatusCode != http.StatusOK {
// 		t.Fatalf("Expected status code %v, got %v", http.StatusOK, resp.StatusCode)
// 	}
//
// 	// Read the response body
// 	respBody, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		t.Fatalf("Failed to read response body: %v", err)
// 	}
//
// 	// Unmarshal the response body into a slice of Document objects
// 	var respDocuments []store.Document
// 	err = json.Unmarshal(respBody, &respDocuments)
// 	if err != nil {
// 		t.Fatalf("Failed to unmarshal response body: %v", err)
// 	}
//
// 	// Check if the returned documents are correct
// 	if len(respDocuments) != 3 {
// 		t.Errorf("Expected 3 documents, got %d", len(respDocuments))
// 	}
//
// 	expectedDocs := []string{"Test Document", "Test Document 2", "Test Document 3"}
// 	for i, doc := range respDocuments {
// 		if doc.DocumentName != expectedDocs[i] {
// 			t.Errorf("Expected document name %s, got %s", expectedDocs[i], doc.DocumentName)
// 		}
// 	}
//
// 	// Cleanup the database
// 	t.Cleanup(func() {
// 		testconfig := config.NewServerConfig()
// 		UserID := testconfig.TestUserID
// 		db.DeleteUserInUsersTable(UserID)
// 		db.DeleteDocument(UserID, "Test Document")
// 		db.DeleteDocument(UserID, "Test Document 2")
// 		db.DeleteDocument(UserID, "Test Document 3")
// 	})
// }
