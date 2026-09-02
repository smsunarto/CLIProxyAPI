package executor

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/runtime/executor/helps"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	"github.com/tidwall/gjson"
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

	fableCloak := newClaudeHeaderTestRequest(t, nil)
	baselineAuth := &cliproxyauth.Auth{
		ID:         auth.ID,
		Attributes: map[string]string{"cloak_mode": "always"},
	}
	err = applyClaudeHeadersWithNativeProfile(
		fableCloak,
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
	assertClaudeFingerprint(t, fableCloak.Header, "claude-cli/2.1.258 (external, sdk-cli)", "0.112.1", "v26.3.0", "MacOS", "arm64")

	legacyCloak := newClaudeHeaderTestRequest(t, nil)
	err = applyClaudeHeadersWithNativeProfile(
		legacyCloak,
		baselineAuth,
		"fable-profile-key",
		false,
		nil,
		[]byte(`{"model":"claude-sonnet-4-6"}`),
		cfg,
		nil,
		false,
		false,
		"11111111-2222-4333-8444-555555555555",
	)
	if err != nil {
		t.Fatalf("applyClaudeHeadersWithNativeProfile() legacy cloak error = %v", err)
	}
	assertClaudeFingerprint(t, legacyCloak.Header, "claude-cli/2.1.220 (external, cli)", "0.94.0", "v26.3.0", "MacOS", "arm64")
	if storedCandidates != 0 {
		t.Fatalf("stored native profile candidates after unconfirmed request = %d, want 0", storedCandidates)
	}
}

func TestApplyCloakingUsesSupportedFable51Version(t *testing.T) {
	payload := []byte(`{"model":"claude-fable-5-1","messages":[{"role":"user","content":"hello"}]}`)
	auth := &cliproxyauth.Auth{Attributes: map[string]string{"cloak_mode": "always"}}

	out, cloaked, err := applyCloaking(context.Background(), nil, auth, payload, "sk-ant-oat-fable", false, true)
	if err != nil {
		t.Fatalf("applyCloaking() error = %v", err)
	}
	if !cloaked {
		t.Fatal("applyCloaking() cloaked = false, want true")
	}
	billing := gjson.GetBytes(out, "system.0.text").String()
	for _, want := range []string{"cc_version=2.1.258.", "cc_entrypoint=sdk-cli"} {
		if !strings.Contains(billing, want) {
			t.Fatalf("billing header = %q, want %q", billing, want)
		}
	}
}
