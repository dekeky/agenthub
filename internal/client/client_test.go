package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agenthub/internal/hub"
)

func TestListAgents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/hub/agents" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":   200,
			"errMsg": "",
			"body": map[string]any{
				"agents": []hub.AgentMeta{{AgentName: "demo", LatestVersion: "1.0.0"}},
				"total":  1,
			},
		})
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL})
	agents, err := c.ListAgents("")
	if err != nil {
		t.Fatal(err)
	}
	if len(agents) != 1 || agents[0].AgentName != "demo" {
		t.Fatalf("unexpected agents: %+v", agents)
	}
}

func TestAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":   404,
			"errMsg": "agent \"missing\" not found",
			"body":   nil,
		})
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL})
	_, err := c.GetAgent("missing")
	if err == nil || err.Error() != `agent "missing" not found` {
		t.Fatalf("unexpected error: %v", err)
	}
}
