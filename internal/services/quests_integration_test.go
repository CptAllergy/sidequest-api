package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sidequest-api/internal/server"
	"testing"
)

func runTestServer() *httptest.Server {
	return httptest.NewServer(server.SetupServer())
}

func TestIntegrationGetQuestsHandler(t *testing.T) {
	testServer := runTestServer()
	defer testServer.Close()

	resp, err := http.Get(fmt.Sprintf("%s/quests", testServer.URL))

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected 200 got: %v", resp.StatusCode)
	}
}
