package registry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func TestTryRefreshModelsCanonicalizesForkModelBeforeCompareAndStore(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"claude": [{
				"id": "claude-fable-5-1",
				"object": "model",
				"type": "claude",
				"context_length": 1000000,
				"max_completion_tokens": 128000,
				"thinking": {
					"min": 1024,
					"max": 128000,
					"zero_allowed": true,
					"dynamic_allowed": false
				}
			}]
		}`))
	}))
	defer server.Close()

	previousURLs := modelsURLs
	modelsURLs = []string{server.URL}
	previousCatalog := getModels()
	refreshCallbackMu.Lock()
	previousCallback := refreshCallback
	previousPending := pendingRefreshChanges
	var notifications [][]string
	refreshCallback = func(changed []string) {
		notifications = append(notifications, append([]string(nil), changed...))
	}
	pendingRefreshChanges = nil
	refreshCallbackMu.Unlock()
	t.Cleanup(func() {
		modelsURLs = previousURLs
		modelsCatalogStore.mu.Lock()
		modelsCatalogStore.data = previousCatalog
		modelsCatalogStore.mu.Unlock()
		refreshCallbackMu.Lock()
		refreshCallback = previousCallback
		pendingRefreshChanges = previousPending
		refreshCallbackMu.Unlock()
	})

	tryRefreshModels(context.Background(), "test Fable refresh")
	tryRefreshModels(context.Background(), "test repeated Fable refresh")

	if len(notifications) != 1 || !slices.Contains(notifications[0], "claude") {
		t.Fatalf("refresh notifications = %#v, want one Claude change", notifications)
	}
	model := LookupModelInfo(claudeFable51ModelID)
	if model == nil || model.Thinking == nil {
		t.Fatalf("stored Fable model = %#v, want thinking metadata", model)
	}
	if model.Thinking.Min != 0 || model.Thinking.Max != 0 || model.Thinking.ZeroAllowed || !model.Thinking.DynamicAllowed {
		t.Fatalf("stored Fable thinking = %#v, want adaptive-only overlay", model.Thinking)
	}
}

func TestClaudeRegistrationRoutesFable51ToClaude(t *testing.T) {
	registry := newTestModelRegistry()
	registry.RegisterClient("claude-auth", "claude", GetClaudeModels())

	providers := registry.GetModelProviders(claudeFable51ModelID)
	if len(providers) != 1 || providers[0] != "claude" {
		t.Fatalf("Fable 5.1 providers = %#v, want [claude]", providers)
	}
}
