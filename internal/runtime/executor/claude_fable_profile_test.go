package executor

import (
	"net/http"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/runtime/executor/helps"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
)

func TestApplyClaudeHeadersPassesMeasuredFable51ProfileWithoutStabilizing(t *testing.T) {
	resetClaudeDeviceProfileCache()
	stabilize := true
	cfg := &config.Config{ClaudeHeaderDefaults: config.ClaudeHeaderDefaults{StabilizeDeviceProfile: &stabilize}}
	auth := &cliproxyauth.Auth{
		ID: "fable-profile-cache-isolation",
		Attributes: map[string]string{
			"header:User-Agent":                  "overridden-client/1.0",
			"header:X-Stainless-Package-Version": "9.9.9",
			"header:X-Stainless-Runtime-Version": "v99.0.0",
		},
	}
	incoming := http.Header{
		"User-Agent":                  {"claude-cli/2.1.258 (external, sdk-cli)"},
		"X-Stainless-Package-Version": {"0.112.1"},
		"X-Stainless-Runtime-Version": {"v26.3.0"},
		"X-Stainless-Os":              {"MacOS"},
		"X-Stainless-Arch":            {"arm64"},
	}

	previousHook := helps.ClaudeDeviceProfileBeforeCandidateStore
	storedCandidates := 0
	helps.ClaudeDeviceProfileBeforeCandidateStore = func(helps.ClaudeDeviceProfile) {
		storedCandidates++
	}
	t.Cleanup(func() {
		helps.ClaudeDeviceProfileBeforeCandidateStore = previousHook
	})

	request := newClaudeHeaderTestRequest(t, incoming)
	err := applyClaudeHeadersWithNativeProfile(
		request,
		auth,
		"fable-profile-key",
		false,
		nil,
		[]byte(`{"model":"claude-fable-5-1"}`),
		cfg,
		incoming,
		true,
		false,
		"11111111-2222-4333-8444-555555555555",
	)
	if err != nil {
		t.Fatalf("applyClaudeHeadersWithNativeProfile() error = %v", err)
	}
	assertClaudeFingerprint(t, request.Header, "claude-cli/2.1.258 (external, sdk-cli)", "0.112.1", "v26.3.0", "MacOS", "arm64")
	if storedCandidates != 0 {
		t.Fatalf("stored native profile candidates = %d, want 0", storedCandidates)
	}

	unconfirmed := newClaudeHeaderTestRequest(t, nil)
	baselineAuth := &cliproxyauth.Auth{
		ID:         auth.ID,
		Attributes: map[string]string{"cloak_mode": "always"},
	}
	err = applyClaudeHeadersWithNativeProfile(
		unconfirmed,
		baselineAuth,
		"fable-profile-key",
		false,
		nil,
		[]byte(`{"model":"claude-fable-5-1"}`),
		cfg,
		nil,
		false,
		false,
		"11111111-2222-4333-8444-555555555555",
	)
	if err != nil {
		t.Fatalf("applyClaudeHeadersWithNativeProfile() unconfirmed error = %v", err)
	}
	assertClaudeFingerprint(t, unconfirmed.Header, "claude-cli/2.1.220 (external, cli)", "0.94.0", "v26.3.0", "MacOS", "arm64")
	if storedCandidates != 0 {
		t.Fatalf("stored native profile candidates after unconfirmed request = %d, want 0", storedCandidates)
	}
}
