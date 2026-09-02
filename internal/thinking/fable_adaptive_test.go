package thinking_test

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/thinking"
	_ "github.com/router-for-me/CLIProxyAPI/v7/internal/thinking/provider/claude"
	"github.com/tidwall/gjson"
)

func TestFable51AutoThinkingUsesAdaptiveWithoutBudget(t *testing.T) {
	body := []byte(`{"thinking":{"type":"adaptive"}}`)
	model := registry.LookupModelInfo("claude-fable-5-1", "claude")
	if model == nil {
		t.Fatal("Claude Fable 5.1 model metadata is missing")
	}
	out, err := thinking.ApplyThinkingWithModelInfo(body, body, model.ID, "claude", "claude", "claude", model)
	if err != nil {
		t.Fatalf("ApplyThinking() error = %v", err)
	}
	if got := gjson.GetBytes(out, "thinking.type").String(); got != "adaptive" {
		t.Fatalf("thinking.type = %q, want adaptive; body=%s", got, out)
	}
	if budget := gjson.GetBytes(out, "thinking.budget_tokens"); budget.Exists() {
		t.Fatalf("thinking.budget_tokens = %s, want absent; body=%s", budget.Raw, out)
	}
}
