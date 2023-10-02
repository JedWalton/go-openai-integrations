package weaviateclient

import (
	"context"
	"errors"
	"log"
	"lucidify-api/modules/config"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

type WeaviateClient interface {
	GetWeaviateClient() *weaviate.Client
	UploadDocument(documentID, userID, name, content string) error
	GetDocument(documentID string) (*Document, error)
	UpdateDocument(documentID, userID, name, content string) error
	DeleteDocument(documentID string) error
	SearchDocumentsByText(limit int, userID string, concepts []string) (*models.GraphQLResponse, error)
}

type WeaviateClientImpl struct {
	client *weaviate.Client
}

type Document struct {
	UserID       string `json:"userId"`
	DocumentName string `json:"documentName"`
	Content      string `json:"content"`
}

func NewWeaviateClient() (WeaviateClient, error) {
	config := config.NewServerConfig()
	cfg := weaviate.Config{
		Host:   "localhost:8090",
		Scheme: "http",
		Headers: map[string]string{
			"X-OpenAI-Api-Key": config.OPENAI_API_KEY,
		},
	}
	client, err := weaviate.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("client is nil after initialization")
	}

	if !classExists(client, "Documents") {
		createWeaviateDocumentsClass(client)
	}

	return &WeaviateClientImpl{client: client}, nil
}

func (w *WeaviateClientImpl) GetWeaviateClient() *weaviate.Client {
	return w.client
}

func createWeaviateDocumentsClass(client *weaviate.Client) {
	if client == nil {
		log.Println("Client is nil in createWeaviateDocumentsClass")
		return
	}

	// Check if the class already exists
	if classExists(client, "Documents") {
		log.Println("Class 'Documents' already exists")
		return
	}

	classObj := &models.Class{
		Class:       "Documents",
		Description: "A document with associated metadata",
		Vectorizer:  "text2vec-openai",
		Properties: []*models.Property{
			{
				DataType:    []string{"string"},
				Description: "Unique identifier of the document",
				Name:        "documentId",
			},
			{
				DataType:    []string{"string"},
				Description: "User identifier associated with the document",
				Name:        "userId",
			},
			{
				DataType:    []string{"string"},
				Description: "Name of the document",
				Name:        "documentName",
			},
			{
				DataType:    []string{"text"},
				Description: "A chunk of the document content",
				Name:        "chunk",
			},
			{
				DataType:    []string{"int"},
				Description: "Unique identifier of the chunk within the document",
				Name:        "chunkId",
			},
			{
				DataType:    []string{"date"},
				Description: "Creation timestamp of the document",
				Name:        "createdAt",
			},
			{
				DataType:    []string{"date"},
				Description: "Update timestamp of the document",
				Name:        "updatedAt",
			},
		},
	}

	err := client.Schema().ClassCreator().WithClass(classObj).Do(context.Background())
	if err != nil {
		panic(err)
	}
}

//	func (w *WeaviateClientImpl) UploadDocument(documentID, userID, name, content string) error {
//		document := map[string]interface{}{
//			"documentId":   documentID,
//			"userId":       userID,
//			"documentName": name,
//			"content":      content,
//		}
//
//		_, err := w.client.Data().Creator().
//			WithID(documentID).
//			WithClassName("Documents").
//			WithProperties(document).
//			Do(context.Background())
//
//		return err
//	}

// Helper function to split content into chunks
func splitContentIntoChunks(content string, chunkSize int) []string {
	var chunks []string
	runes := []rune(content)

	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}

	return chunks
}

func (w *WeaviateClientImpl) UploadDocument(documentID, userID, name, content string) error {
	// Split the content into chunks
	chunkSize := 1000
	chunks := splitContentIntoChunks(content, chunkSize)

	for i, chunk := range chunks {
		document := map[string]interface{}{
			"documentId":   documentID,
			"userId":       userID,
			"documentName": name,
			"chunk":        chunk,
			"chunkId":      i,
		}

		_, err := w.client.Data().Creator().
			WithID(documentID).
			WithClassName("Documents").
			WithProperties(document).
			Do(context.Background())

		if err != nil {
			return err
		}
	}

	return nil
}

func (w *WeaviateClientImpl) GetDocument(documentID string) (*Document, error) {
	objects, err := w.client.Data().ObjectsGetter().
		WithClassName("Documents").
		WithID(documentID).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	// If no objects are returned, return an error
	if len(objects) == 0 {
		return nil, errors.New("no documents found")
	}

	// Combine chunks to form the complete document content
	var content string
	for _, obj := range objects {
		if obj.Properties == nil {
			return nil, errors.New("properties does not exist")
		}

		chunkValue, exists := obj.Properties.(map[string]interface{})["chunk"]
		if !exists || chunkValue == nil {
			return nil, errors.New("chunk does not exist")
		}
		content += chunkValue.(string)
	}

	// Assume the first object is the one you're looking for
	obj := objects[0]

	// Additional checks for each field before type assertion
	userID, ok := obj.Properties.(map[string]interface{})["userId"]
	if !ok || userID == nil {
		return nil, errors.New("userId does not exist")
	}

	documentName, ok := obj.Properties.(map[string]interface{})["documentName"]
	if !ok || documentName == nil {
		return nil, errors.New("documentName does not exist")
	}

	// Convert the object to a Document
	doc := &Document{
		UserID:       userID.(string),
		DocumentName: documentName.(string),
		Content:      content,
	}

	return doc, nil
}

//	func (w *WeaviateClientImpl) UpdateDocumentContent(documentID, content string) error {
//		document := map[string]interface{}{
//			"content": content,
//		}
//
//		err := w.client.Data().Updater().
//			WithMerge().
//			WithID(documentID).
//			WithClassName("Documents").
//			WithProperties(document).
//			Do(context.Background())
//
//		return err
//	}
func (w *WeaviateClientImpl) UpdateDocument(documentID, userID, name, content string) error {
	// First, delete all existing chunks for the document
	err := w.client.Data().Deleter().
		WithClassName("Documents").
		WithID(documentID).
		Do(context.Background())
	if err != nil {
		return err
	}

	// Now, use the UploadDocument function to add the new content
	err = w.UploadDocument(documentID, userID, name, content)
	if err != nil {
		return err
	}
	return nil
}

// func (w *WeaviateClientImpl) UpdateDocumentName(documentID, documentName string) error {
// 	document := map[string]interface{}{
// 		"documentName": documentName,
// 	}
//
// 	err := w.client.Data().Updater().
// 		WithMerge().
// 		WithID(documentID).
// 		WithClassName("Documents").
// 		WithProperties(document).
// 		Do(context.Background())
//
// 	return err
// }

func (w *WeaviateClientImpl) DeleteDocument(documentID string) error {
	err := w.client.Data().Deleter().
		WithClassName("Documents").
		WithID(documentID).
		Do(context.Background())

	return err
}

func classExists(client *weaviate.Client, className string) bool {
	schema, err := client.Schema().ClassGetter().WithClassName(className).Do(context.Background())
	if err != nil {
		return false
	}
	log.Printf("%v", schema)
	return true
}

func (w *WeaviateClientImpl) SearchDocumentsByText(limit int, userID string, concepts []string) (*models.GraphQLResponse, error) {
	className := "Documents"

	documentName := graphql.Field{Name: "documentName"}
	content := graphql.Field{Name: "content"}
	_additional := graphql.Field{
		Name: "_additional", Fields: []graphql.Field{
			{Name: "certainty"}, // only supported if distance==cosine
			{Name: "distance"},  // always supported
		},
	}

	distance := float32(0.6)
	// moveAwayFrom := &graphql.MoveParameters{
	// 	Concepts: []string{"finance"},
	// 	Force:    0.45,
	// }
	// moveTo := &graphql.MoveParameters{
	// 	Concepts: []string{"haute couture"},
	// 	Force:    0.85,
	// }
	nearText := w.client.GraphQL().NearTextArgBuilder().
		WithConcepts(concepts).
		WithDistance(distance) // use WithCertainty(certainty) prior to v1.14
		// WithMoveTo(moveTo).
		// WithMoveAwayFrom(moveAwayFrom)

		// Creating the where filter
	whereFilter := filters.Where().
		WithPath([]string{"userId"}).
		WithOperator(filters.Equal).
		WithValueText(userID)

	ctx := context.Background()

	result, err := w.client.GraphQL().Get().
		WithClassName(className).
		WithFields(documentName, content, _additional).
		WithNearText(nearText).
		WithLimit(limit).
		WithWhere(whereFilter).
		Do(ctx)

	if err != nil {
		panic(err)
	}
	return result, nil
}
