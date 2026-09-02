package helps

import (
	"net/http"
	"strings"
)

type claudeNativeSoftwareProfile struct {
	UserAgent      string
	PackageVersion string
	RuntimeVersion string
	cloakModels    []string
}

var forkClaudeNativeSoftwareProfiles = []claudeNativeSoftwareProfile{
	{
		UserAgent:      "claude-cli/2.1.258 (external, sdk-cli)",
		PackageVersion: "0.112.1",
		RuntimeVersion: "v26.3.0",
		cloakModels:    []string{"claude-fable-5-1"},
	},
}

func claudeNativeSoftwareProfileForCloakModel(model string) (claudeNativeSoftwareProfile, bool) {
	model = strings.ToLower(strings.TrimSpace(model))
	if suffix := strings.LastIndex(model, "("); suffix >= 0 && strings.HasSuffix(model, ")") {
		model = strings.TrimSpace(model[:suffix])
	}
	for _, profile := range forkClaudeNativeSoftwareProfiles {
		for _, candidate := range profile.cloakModels {
			if model == candidate {
				return profile, true
			}
		}
	}
	return claudeNativeSoftwareProfile{}, false
}

func matchClaudeNativeSoftwareProfile(headers http.Header) (claudeNativeSoftwareProfile, bool) {
	if headers == nil {
		return claudeNativeSoftwareProfile{}, false
	}

	for _, profile := range forkClaudeNativeSoftwareProfiles {
		if claudeHeaderHasExactValue(headers, "User-Agent", profile.UserAgent) &&
			claudeHeaderHasExactValue(headers, "X-Stainless-Package-Version", profile.PackageVersion) &&
			claudeHeaderHasExactValue(headers, "X-Stainless-Runtime-Version", profile.RuntimeVersion) {
			return profile, true
		}
	}
	return claudeNativeSoftwareProfile{}, false
}

func claudeHeaderHasExactValue(headers http.Header, name, want string) bool {
	values := make([]string, 0, 1)
	for key, candidates := range headers {
		if strings.EqualFold(key, name) {
			values = append(values, candidates...)
		}
	}
	return len(values) == 1 && values[0] == want
}

// IsClaudeNativeSoftwareProfile reports whether all measured software headers
// match one fork-owned native profile exactly.
func IsClaudeNativeSoftwareProfile(headers http.Header) bool {
	_, ok := matchClaudeNativeSoftwareProfile(headers)
	return ok
}

// ApplyClaudeNativeSoftwareProfileHeaders copies a recognized native software
// tuple without involving the synthetic device-profile cache.
func ApplyClaudeNativeSoftwareProfileHeaders(r *http.Request, headers http.Header) bool {
	if r == nil {
		return false
	}
	profile, ok := matchClaudeNativeSoftwareProfile(headers)
	if !ok {
		return false
	}
	r.Header.Set("User-Agent", profile.UserAgent)
	r.Header.Set("X-Stainless-Package-Version", profile.PackageVersion)
	r.Header.Set("X-Stainless-Runtime-Version", profile.RuntimeVersion)
	return true
}
