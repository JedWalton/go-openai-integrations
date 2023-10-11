// //go:build integration
// // +build integration
package store

import (
	"log"
	"lucidify-api/modules/store/postgresqlclient"
	"lucidify-api/modules/store/storemodels"
	"lucidify-api/modules/store/weaviateclient"
	"os"
	"testing"

	"github.com/google/uuid"
)

func createTestUserInDb() string {
	db, err := postgresqlclient.NewPostgreSQL()

	// the user id registered by the jwt token must exist in the local database
	user := postgresqlclient.User{
		UserID:           "TestDocumentsServiceIntegrationTestUUID",
		ExternalID:       "TestDocumentsServiceIntegrationTestExternalID",
		Username:         "TestDocumentsServiceIntegrationTestUsername",
		PasswordEnabled:  true,
		Email:            "TestDocumentServiceIntTest@gmail.com",
		FirstName:        "TestDocumentsServiceIntegrationTestFirstName",
		LastName:         "TestDocumentsServiceIntegrationTestLastName",
		ImageURL:         "https://TestDocumentsServiceIntegrationTestURL.com/image.jpg",
		ProfileImageURL:  "https://TestDocumentServiceTestProfileURL.com/profile.jpg",
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
	return user.UserID
}

func readFileContent(filename string) (string, error) {
	contentBytes, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(contentBytes), nil
}

func TestSplitContentIntoChunks(t *testing.T) {
	// Define a struct for test cases
	type testCase struct {
		filename       string
		expectedChunks int
	}

	// Create a slice of test cases
	testCases := []testCase{
		{"test_doc_user1_01.txt", 4},
		{"test_doc_cats.txt", 4},
		{"test_doc_vector_databases.txt", 4},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			content, err := readFileContent(tc.filename)
			if err != nil {
				t.Errorf("failed to read file content: %v", err)
			}

			document := storemodels.Document{
				DocumentUUID: uuid.New(),
				UserID:       "TestStoreIntegrationTestUserUUID",
				DocumentName: "test_document_name",
				Content:      content,
			}

			// Use the function to split the content
			chunks, err := splitContentIntoChunks(document)
			if err != nil {
				t.Errorf("failed to split content: %v", err)
			}
			if len(chunks) != tc.expectedChunks {
				t.Errorf("incorrect number of chunks: got %v, want %v", len(chunks), tc.expectedChunks)
			}
		})
	}
}

func TestUploadDocumentIntegration(t *testing.T) {
	// 1. Setup
	// Initialize PostgreSQL for tests
	db, err := postgresqlclient.NewPostgreSQL()
	if err != nil {
		t.Fatalf("Failed to initialize PostgreSQL: %v", err)
	}

	// Initialize Weaviate for tests
	weaviateClient, err := weaviateclient.NewWeaviateClientTest()
	if err != nil {
		t.Fatalf("Failed to create Weaviate client: %v", err)
	}

	documentService := NewDocumentService(db, weaviateClient)

	// Test data
	name := "test-document-name"
	content := "This is a test document content."

	user := postgresqlclient.User{
		UserID:           "documents_service_integration_test_user_id",
		ExternalID:       "documents_service_external_ID",
		Username:         "TestDocumentsServiceIntegrationTableUsername",
		PasswordEnabled:  true,
		Email:            "TestDocumentsService@example.com",
		FirstName:        "TestDocumentsCreateUserInUsersTableCreateTest",
		LastName:         "TestDocumentsCreateUserInUsersTableUser",
		ImageURL:         "https://TestCreateUserInUsersTable.com/image.jpg",
		ProfileImageURL:  "https://TestCreateUserInUsersTable.com/profile.jpg",
		TwoFactorEnabled: false,
		CreatedAt:        1654012591514,
		UpdatedAt:        1654012591514,
	}

	err = db.CreateUserInUsersTable(user)
	if err != nil {
		t.Errorf("Failed to create user: %v", err)
	}

	// 2. Call the function
	document, err := documentService.UploadDocument(user.UserID, name, content)
	if err != nil {
		t.Fatalf("Failed to upload document: %v", err)
	}

	// // 3. Verify upload
	doc, err := db.GetDocumentByUUID(document.DocumentUUID)
	if err != nil || doc == nil {
		t.Error("Document was not uploaded to PostgreSQL")
	}

	chunks, err := db.GetChunksOfDocument(document)
	if err != nil || len(chunks) == 0 {
		t.Error("Chunks were not uploaded to PostgreSQL")
	}

	err = db.DeleteDocumentByUUID(document.DocumentUUID)
	if err != nil {
		t.Errorf("Failed to delete test document: %v", err)
	}

	chunks, err = db.GetChunksOfDocument(document)
	if err != nil || len(chunks) != 0 {
		t.Error("Chunks were not uploaded to PostgreSQL")
	}

	chunksFromWeaviate, err := weaviateClient.GetChunks(chunks)
	if err != nil || len(chunksFromWeaviate) != 0 {
		t.Error("Chunks were not uploaded to Weaviate")
	}

	for i, chunk := range chunksFromWeaviate {
		if chunk.ChunkID != chunks[i].ChunkID {
			t.Error("Chunks ChunkID inconsistent before and after uploading chunks to weaviate")
		}
		if chunk.UserID != chunks[i].UserID {
			t.Error("Chunks UserID inconsistent before and after uploading chunks to weaviate")
		}
		if chunk.DocumentID != chunks[i].DocumentID {
			t.Error("Chunks DocumentID inconsistent before and after uploading chunks to weaviate")
		}
		if chunk.ChunkContent != chunks[i].ChunkContent {
			t.Error("Chunks ChunkContent are inconsistent before and after uploading chunks to weaviate")
		}
		if chunk.ChunkIndex != chunks[i].ChunkIndex {
			t.Error("Chunks ChunkIndex are inconsistent before and after uploading chunks to weaviate")
		}
	}

	document, err = documentService.UploadDocument(user.UserID, name, content)
	if err != nil {
		t.Fatalf("Failed to upload document: %v", err)
	}
	doc, err = db.GetDocumentByUUID(document.DocumentUUID)
	if err != nil || doc == nil {
		t.Error("Document was not uploaded to PostgreSQL")
	}

	chunks, err = db.GetChunksOfDocument(document)
	if err != nil || len(chunks) == 0 {
		t.Error("Chunks were not uploaded to PostgreSQL")
	}

	chunksFromWeaviate, err = weaviateClient.GetChunks(chunks)
	if err != nil || len(chunksFromWeaviate) == 0 {
		t.Error("Chunks were not uploaded to Weaviate")
	}

	err = db.DeleteUserInUsersTable(user.UserID)
	if err != nil {
		t.Errorf("failed to delete test user: %v", err)
	}
	doc, err = db.GetDocumentByUUID(document.DocumentUUID)
	if err == nil || doc != nil {
		t.Error("Document not deleted from PostgreSQL after user deleted.")
	}
	chunks, err = db.GetChunksOfDocument(doc)
	if err == nil || len(chunks) != 0 {
		t.Error("Chunks not deleted PostgreSQL after user deleted.")
	}

	chunksFromWeaviate, err = weaviateClient.GetChunks(chunks)
	if err != nil || len(chunksFromWeaviate) != 0 {
		t.Error("Chunks not deleted from Weaviate after user deleted.")
	}

	err = db.CreateUserInUsersTable(user)
	if err != nil {
		t.Errorf("Failed to create user: %v", err)
	}

	name2 := "test-document-name2"
	content2 := "This is a test document content2."

	document, err = documentService.UploadDocument(user.UserID, name, content)
	if err != nil {
		t.Fatalf("Failed to upload document: %v", err)
	}

	document, err = documentService.UploadDocument(user.UserID, name2, content2)
	if err != nil {
		t.Fatalf("Failed to upload document: %v", err)
	}

	allDocs, err := documentService.GetAllDocuments(user.UserID)
	if err != nil || len(allDocs) != 2 {
		t.Error("Document was not uploaded to PostgreSQL")
	}

	// 4. Cleanup
	// Execute cleanup tasks after all checks
	t.Cleanup(func() {
		err = db.DeleteUserInUsersTable(user.UserID)
		if err != nil {
			t.Errorf("failed to delete test user: %v", err)
		}
	})
}
