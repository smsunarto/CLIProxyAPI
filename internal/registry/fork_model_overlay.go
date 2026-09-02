package registry

const claudeFable51ModelID = "claude-fable-5-1"

func adaptiveOnlyThinking() *ThinkingSupport {
	return &ThinkingSupport{
		DynamicAllowed: true,
		Levels:         []string{"low", "medium", "high", "xhigh", "max"},
	}
}

func forkClaudeFable51Model() *ModelInfo {
	return &ModelInfo{
		ID:                       claudeFable51ModelID,
		Object:                   "model",
		Created:                  1788220800,
		OwnedBy:                  "anthropic",
		Type:                     "claude",
		DisplayName:              "Claude Fable 5.1",
		Description:              "Anthropic's most capable widely released model, for the most demanding reasoning and long-horizon agentic work",
		ContextLength:            1_000_000,
		MaxCompletionTokens:      128_000,
		SupportedInputModalities: []string{"text", "image"},
		SupportedOutputModalities: []string{
			"text",
		},
		Thinking: adaptiveOnlyThinking(),
	}
}

func applyForkModelOverlay(catalog *staticModelsJSON) {
	if catalog == nil {
		return
	}

	fable := forkClaudeFable51Model()
	found := false
	for index, model := range catalog.Claude {
		if model == nil || model.ID != claudeFable51ModelID {
			continue
		}
		catalog.Claude[index] = fable
		found = true
	}
	if !found {
		catalog.Claude = append(catalog.Claude, fable)
	}
}
