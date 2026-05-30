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

func TestGetAgentWithVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/hub/agents/demo" || r.URL.Query().Get("version") != "1.0.0" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 200, "errMsg": "",
			"body": hub.AgentDetail{AgentMeta: hub.AgentMeta{AgentName: "demo", LatestVersion: "1.0.0"}},
		})
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL})
	detail, err := c.GetAgent("demo", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if detail.AgentName != "demo" {
		t.Fatalf("unexpected detail: %+v", detail)
	}
}

func TestUpdateAgentMeta(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/hub/agents/demo" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Fatalf("auth header = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 200, "errMsg": "",
			"body": hub.AgentMeta{AgentName: "demo", DisplayName: "Demo"},
		})
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, Token: "secret"})
	meta, err := c.UpdateAgentMeta("demo", UpdateMetaRequest{DisplayName: "Demo"})
	if err != nil {
		t.Fatal(err)
	}
	if meta.DisplayName != "Demo" {
		t.Fatalf("unexpected meta: %+v", meta)
	}
}

func TestDeleteAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/hub/agents/demo" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 200, "errMsg": "",
			"body": map[string]any{"deleted": true},
		})
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, Token: "secret"})
	if err := c.DeleteAgent("demo"); err != nil {
		t.Fatal(err)
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
	_, err := c.GetAgent("missing", "")
	if err == nil || err.Error() != `agent "missing" not found` {
		t.Fatalf("unexpected error: %v", err)
	}
}
