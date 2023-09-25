package documents

import (
	"lucidify-api/modules/config"
	"lucidify-api/modules/middleware"
	"lucidify-api/modules/store"
	"net/http"

	"github.com/clerkinc/clerk-sdk-go/clerk"
)

func SetupRoutes(config *config.ServerConfig, mux *http.ServeMux, storeInstance *store.Store, client clerk.Client) *http.ServeMux {
	mux = SetupDocumentsUploadHandler(config, mux, storeInstance, client)
	mux = SetupDocumentsGetDocumentHandler(config, mux, storeInstance, client)
	mux = SetupDocumentsGetAllDocumentHandler(config, mux, storeInstance, client)
	mux = SetupDocumentsDeleteDocumentHandler(config, mux, storeInstance, client)
	mux = SetupDocumentsUpdateDocumentHandler(config, mux, storeInstance, client)

	return mux
}

func SetupDocumentsUploadHandler(config *config.ServerConfig, mux *http.ServeMux, storeInstance *store.Store, client clerk.Client) *http.ServeMux {

	handler := DocumentsUploadHandler(storeInstance, client)

	injectActiveSession := clerk.WithSession(client)

	handler = middleware.CORSMiddleware(config.AllowedOrigins)(handler)
	handler = middleware.Logging(handler)

	mux.Handle("/documents/upload", injectActiveSession(handler))

	return mux
}

func SetupDocumentsGetDocumentHandler(config *config.ServerConfig, mux *http.ServeMux, storeInstance *store.Store, client clerk.Client) *http.ServeMux {

	handler := DocumentsGetDocumentHandler(storeInstance, client)

	injectActiveSession := clerk.WithSession(client)

	handler = middleware.CORSMiddleware(config.AllowedOrigins)(handler)
	handler = middleware.Logging(handler)

	mux.Handle("/documents/getdocument", injectActiveSession(handler))

	return mux
}

func SetupDocumentsGetAllDocumentHandler(config *config.ServerConfig, mux *http.ServeMux, storeInstance *store.Store, client clerk.Client) *http.ServeMux {

	handler := DocumentsGetAllDocumentsHandler(storeInstance, client)

	injectActiveSession := clerk.WithSession(client)

	handler = middleware.CORSMiddleware(config.AllowedOrigins)(handler)
	handler = middleware.Logging(handler)

	mux.Handle("/documents/getalldocuments", injectActiveSession(handler))

	return mux
}

func SetupDocumentsDeleteDocumentHandler(config *config.ServerConfig, mux *http.ServeMux, storeInstance *store.Store, client clerk.Client) *http.ServeMux {

	handler := DocumentsDeleteDocumentHandler(storeInstance, client)

	injectActiveSession := clerk.WithSession(client)

	handler = middleware.CORSMiddleware(config.AllowedOrigins)(handler)
	handler = middleware.Logging(handler)

	mux.Handle("/documents/deletedocument", injectActiveSession(handler))

	return mux
}

func SetupDocumentsUpdateDocumentHandler(config *config.ServerConfig, mux *http.ServeMux, storeInstance *store.Store, client clerk.Client) *http.ServeMux {

	handler := DocumentsUpdateDocumentHandler(storeInstance, client)

	injectActiveSession := clerk.WithSession(client)

	handler = middleware.CORSMiddleware(config.AllowedOrigins)(handler)
	handler = middleware.Logging(handler)

	mux.Handle("/documents/updatedocument", injectActiveSession(handler))

	return mux
}
