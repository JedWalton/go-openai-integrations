//go:build integration
// +build integration

package clerk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"lucidify-api/modules/config"
	"lucidify-api/modules/store"
	"os/exec"
	"testing"
	"time"
)

func checkUserInDB(db *store.Store, userID string, retries int) error {
	for i := 0; i < retries; i++ {
		_, err := db.GetUser(userID)
		if err == nil {
			return nil
		}
		time.Sleep(time.Second) // Wait for 1 second before retrying
	}
	return fmt.Errorf("User not found after %d retries", retries)
}

func TestIntegration_usercreatedevent(t *testing.T) {
	testconfig := config.NewTestServerConfig()
	db := testconfig.TestStore

	MakeCurlRequest := func() (string, error) {
		cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/clerk/webhook", "-H", "Content-Type: application/json", "-d", "@test/example_user_created_event.txt")
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	}

	response, err := MakeCurlRequest()
	if err != nil {
		t.Fatalf("Failed to make curl request: %v", err)
	}
	log.Printf("Response: %s\n", response)

	content, err := ioutil.ReadFile("test/example_user_created_event.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Parse the content
	var event ClerkEvent
	err = json.Unmarshal(content, &event)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// Now you can use the event variable
	UserID := event.Data["id"].(string)

	err = checkUserInDB(db, UserID, 5) // Try 5 times
	if err != nil {
		t.Fatalf("Failed to fetch user, user not in db with UserID: %v", err)
	}

	db.DeleteUser(UserID)
}

func TestIntegration_userupdatedevent(t *testing.T) {
	testconfig := config.NewTestServerConfig()
	db := testconfig.TestStore

	MakeCurlRequest := func() (string, error) {
		cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/clerk/webhook", "-H", "Content-Type: application/json", "-d", "@test/example_user_created_event.txt")
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	}

	_, err := MakeCurlRequest()
	if err != nil {
		t.Fatalf("Failed to make curl request: %v", err)
	}

	MakeCurlRequest = func() (string, error) {
		cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/clerk/webhook", "-H", "Content-Type: application/json", "-d", "@test/example_user_updated_event.txt")
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	}

	_, err = MakeCurlRequest()
	if err != nil {
		t.Fatalf("Failed to make curl request: %v", err)
	}

	content, err := ioutil.ReadFile("test/example_user_updated_event.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	// Parse the content
	var event ClerkEvent
	err = json.Unmarshal(content, &event)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	UserID := event.Data["id"].(string)
	err = checkUserInDB(db, UserID, 5) // Try 5 times
	if err != nil {
		t.Fatalf("Failed to fetch user, user not in db with UserID: %v", err)
	}

	User, err := db.GetUser(UserID)
	if err != nil {
		t.Fatalf("Failed to fetch user: %v", err)
	}
	username := User.Username
	if username != "example_username_updated" {
		t.Fatalf("Expected username to be testuser, got: %s", username)
	}

	db.DeleteUser(UserID)
}

func checkUserDeletedFromDB(db *store.Store, userID string, retries int) error {
	for i := 0; i < retries; i++ {
		_, err := db.GetUser(userID)
		if err != nil {
			// If the user is not found, it means the user has been deleted
			return nil
		}
		time.Sleep(time.Second) // Wait for 1 second before retrying
	}
	return fmt.Errorf("User still exists in the database after %d retries", retries)
}

func TestIntegration_userdeletedevent(t *testing.T) {
	testconfig := config.NewTestServerConfig()
	db := testconfig.TestStore

	MakeCurlRequest := func() (string, error) {
		cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/clerk/webhook", "-H", "Content-Type: application/json", "-d", "@test/example_user_created_event.txt")
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	}

	response, err := MakeCurlRequest()
	if err != nil {
		t.Fatalf("Failed to make curl request: %v", err)
	}
	log.Printf("Response: %s\n", response)

	content, err := ioutil.ReadFile("test/example_user_created_event.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Parse the content
	var event ClerkEvent
	err = json.Unmarshal(content, &event)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// Now you can use the event variable
	UserID := event.Data["id"].(string)

	err = checkUserInDB(db, UserID, 5) // Try 5 times
	if err != nil {
		t.Fatalf("Failed to fetch user, user not in db with UserID: %v", err)
	}

	MakeCurlRequest = func() (string, error) {
		cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/clerk/webhook", "-H", "Content-Type: application/json", "-d", "@test/example_user_deleted_event.txt")
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	}

	_, err = MakeCurlRequest()

	err = checkUserDeletedFromDB(db, UserID, 5) // Try 5 times

	db.DeleteUser(UserID)
}
