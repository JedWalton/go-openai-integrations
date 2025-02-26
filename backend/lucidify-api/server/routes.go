package server

import (
	"lucidify-api/data/store/postgresqlclient"
	"lucidify-api/data/store/weaviateclient"
	"lucidify-api/http/chatapi"
	"lucidify-api/http/clerkapi"
	"lucidify-api/http/documentsapi"
	"lucidify-api/http/syncapi"
	"lucidify-api/server/config"
	"lucidify-api/service/chatservice"
	"lucidify-api/service/documentservice"
	"lucidify-api/service/syncservice"
	"lucidify-api/service/userservice"
	"net/http"

	"github.com/clerkinc/clerk-sdk-go/clerk"
)

func SetupRoutes(
	config *config.ServerConfig,
	mux *http.ServeMux,
	storeInstance *postgresqlclient.PostgreSQL,
	clerkInstance clerk.Client,
	weaviateInstance weaviateclient.WeaviateClient,
	documentsService documentservice.DocumentService,
	cvs chatservice.ChatVectorService,
	syncService syncservice.SyncService,
	userService userservice.UserService) {

	chatapi.SetupRoutes(config, mux, cvs, clerkInstance)
	documentsapi.SetupRoutes(config, mux, documentsService, clerkInstance)
	clerkapi.SetupRoutes(storeInstance, userService, config, mux)
	syncapi.SetupRoutes(config, mux, clerkInstance, syncService)
}
