package syncapi

import (
	"encoding/json"
	"io"
	"log"
	"lucidify-api/service/syncservice"
	"net/http"
)

// LocalStorageKey defines valid keys for LocalStorage operations.
type LocalStorageKey string

const (
	apiKey               LocalStorageKey = "apiKey"
	conversationHistory  LocalStorageKey = "conversationHistory"
	selectedConversation LocalStorageKey = "selectedConversation"
	theme                LocalStorageKey = "theme"
	folders              LocalStorageKey = "folders"
	prompts              LocalStorageKey = "prompts"
	showChatbar          LocalStorageKey = "showChatbar"
	showPromptbar        LocalStorageKey = "showPromptbar"
	pluginKeys           LocalStorageKey = "pluginKeys"
	settings             LocalStorageKey = "settings"
)

// IsValid checks if the provided key is a valid LocalStorageKey.
func (key LocalStorageKey) IsValid() bool {
	switch key {
	case apiKey, conversationHistory, selectedConversation, theme, folders,
		prompts, showChatbar, showPromptbar, pluginKeys, settings:
		return true
	}
	return false
}

func MethodNotAllowed(w http.ResponseWriter) {
	response := syncservice.ServerResponse{
		Success: false,
		Message: "Method not allowed",
	}
	sendJSONResponse(w, http.StatusMethodNotAllowed, response)
}

// This is a utility function to send JSON responses
func sendJSONResponse(w http.ResponseWriter, statusCode int, response syncservice.ServerResponse) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func SyncHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		key := r.URL.Query().Get("key")

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("Error reading request body:", err)
			// Handle the error, maybe return a response indicating the error.
			return
		}
		value := string(bodyBytes)

		var response syncservice.ServerResponse

		switch r.Method {
		case http.MethodGet:
			resp := syncservice.HandleGet(key)
			response = resp
			if response.Success {
				response.Data = resp.Data
			}
		case http.MethodDelete:
			response = syncservice.HandleRemove(key)
		case http.MethodPost:
			response = syncservice.HandleSet(key, value)
		default:
			response = syncservice.ServerResponse{
				Success: false,
				Message: "Method not allowed",
			}
			sendJSONResponse(w, http.StatusMethodNotAllowed, response)
			return
		}
		sendJSONResponse(w, http.StatusOK, response)
	}
}
