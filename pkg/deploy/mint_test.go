package deploy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateConfigHash(t *testing.T) {
	tests := []struct {
		name     string
		config   *AgentConfig
		wantSame bool // if true, compare with second config
		config2  *AgentConfig
	}{
		{
			name: "basic config generates hash",
			config: &AgentConfig{
				AgentID:   "test-agent",
				Name:      "Test Agent",
				AgentType: "command",
				Capabilities: []Capability{
					{Name: "cap1", Description: "desc1"},
				},
				Categories:  []string{"AI"},
				NlpFallback: false,
			},
		},
		{
			name: "same config generates same hash",
			config: &AgentConfig{
				AgentID:   "test-agent",
				Name:      "Test Agent",
				AgentType: "command",
				Capabilities: []Capability{
					{Name: "cap1"},
				},
				Categories: []string{"AI"},
			},
			wantSame: true,
			config2: &AgentConfig{
				AgentID:   "test-agent",
				Name:      "Test Agent",
				AgentType: "command",
				Capabilities: []Capability{
					{Name: "cap1"},
				},
				Categories: []string{"AI"},
			},
		},
		{
			name: "different agent_id generates different hash",
			config: &AgentConfig{
				AgentID:      "agent-1",
				Name:         "Test",
				AgentType:    "command",
				Capabilities: []Capability{{Name: "cap"}},
				Categories:   []string{"AI"},
			},
			wantSame: false,
			config2: &AgentConfig{
				AgentID:      "agent-2",
				Name:         "Test",
				AgentType:    "command",
				Capabilities: []Capability{{Name: "cap"}},
				Categories:   []string{"AI"},
			},
		},
		{
			name: "capabilities order doesnt matter",
			config: &AgentConfig{
				AgentID:   "test",
				Name:      "Test",
				AgentType: "command",
				Capabilities: []Capability{
					{Name: "aaa"},
					{Name: "zzz"},
				},
				Categories: []string{"AI"},
			},
			wantSame: true,
			config2: &AgentConfig{
				AgentID:   "test",
				Name:      "Test",
				AgentType: "command",
				Capabilities: []Capability{
					{Name: "zzz"},
					{Name: "aaa"},
				},
				Categories: []string{"AI"},
			},
		},
		{
			name: "categories order doesnt matter",
			config: &AgentConfig{
				AgentID:      "test",
				Name:         "Test",
				AgentType:    "command",
				Capabilities: []Capability{{Name: "cap"}},
				Categories:   []string{"AI", "Automation"},
			},
			wantSame: true,
			config2: &AgentConfig{
				AgentID:      "test",
				Name:         "Test",
				AgentType:    "command",
				Capabilities: []Capability{{Name: "cap"}},
				Categories:   []string{"Automation", "AI"},
			},
		},
		{
			name: "command parameters affect hash",
			config: &AgentConfig{
				AgentID:      "test",
				Name:         "Test",
				Description:  "Test description",
				AgentType:    "command",
				Capabilities: []Capability{{Name: "cap", Description: "desc"}},
				Categories:   []string{"AI"},
				Commands: []Command{{
					Trigger:      "cmd1",
					PricePerUnit: 0.01,
					PriceType:    "task-transaction",
					TaskUnit:     "per-query",
				}},
			},
			wantSame: false,
			config2: &AgentConfig{
				AgentID:      "test",
				Name:         "Test",
				Description:  "Test description",
				AgentType:    "command",
				Capabilities: []Capability{{Name: "cap", Description: "desc"}},
				Categories:   []string{"AI"},
				Commands: []Command{{
					Trigger:      "cmd1",
					Argument:     "<input>",
					Description:  "Does something",
					PricePerUnit: 0.01,
					PriceType:    "task-transaction",
					TaskUnit:     "per-query",
					Parameters: []CommandParameter{
						{Name: "input", Type: "string", Required: true, Description: "The input"},
					},
				}},
			},
		},
		{
			name: "command strictArg affects hash",
			config: &AgentConfig{
				AgentID:      "test",
				Name:         "Test",
				Description:  "Test description",
				AgentType:    "command",
				Capabilities: []Capability{{Name: "cap", Description: "desc"}},
				Categories:   []string{"AI"},
				Commands: []Command{{
					Trigger:      "cmd1",
					PricePerUnit: 0.01,
					PriceType:    "task-transaction",
					TaskUnit:     "per-query",
				}},
			},
			wantSame: false,
			config2: &AgentConfig{
				AgentID:      "test",
				Name:         "Test",
				Description:  "Test description",
				AgentType:    "command",
				Capabilities: []Capability{{Name: "cap", Description: "desc"}},
				Categories:   []string{"AI"},
				Commands: []Command{{
					Trigger:      "cmd1",
					PricePerUnit: 0.01,
					PriceType:    "task-transaction",
					TaskUnit:     "per-query",
					StrictArg:    boolPtr(true),
					MinArgs:      intPtr(1),
					MaxArgs:      intPtr(1),
				}},
			},
		},
		{
			name: "capability description affects hash",
			config: &AgentConfig{
				AgentID:      "test",
				Name:         "Test",
				Description:  "Test description",
				AgentType:    "command",
				Capabilities: []Capability{{Name: "cap", Description: "original desc"}},
				Categories:   []string{"AI"},
			},
			wantSame: false,
			config2: &AgentConfig{
				AgentID:      "test",
				Name:         "Test",
				Description:  "Test description",
				AgentType:    "command",
				Capabilities: []Capability{{Name: "cap", Description: "updated desc"}},
				Categories:   []string{"AI"},
			},
		},
		{
			name: "description affects hash",
			config: &AgentConfig{
				AgentID:      "test",
				Name:         "Test",
				Description:  "Description 1",
				AgentType:    "command",
				Capabilities: []Capability{{Name: "cap"}},
				Categories:   []string{"AI"},
			},
			wantSame: false,
			config2: &AgentConfig{
				AgentID:      "test",
				Name:         "Test",
				Description:  "Totally different description",
				AgentType:    "command",
				Capabilities: []Capability{{Name: "cap"}},
				Categories:   []string{"AI"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := GenerateConfigHash(tt.config)

			// Basic checks
			if hash1 == "" {
				t.Error("GenerateConfigHash returned empty string")
			}
			if len(hash1) != 64 {
				t.Errorf("GenerateConfigHash returned hash of length %d, want 64", len(hash1))
			}

			// Compare with second config if provided
			if tt.config2 != nil {
				hash2 := GenerateConfigHash(tt.config2)
				if tt.wantSame && hash1 != hash2 {
					t.Errorf("Expected same hash, got different:\n  hash1: %s\n  hash2: %s", hash1, hash2)
				}
				if !tt.wantSame && hash1 == hash2 {
					t.Errorf("Expected different hash, got same: %s", hash1)
				}
			}
		})
	}
}

func TestGenerateConfigHashNormalizesLegacyChoiceParameters(t *testing.T) {
	legacyJSON := []byte(`{
		"name": "Test Agent",
		"agentId": "test-agent",
		"description": "A valid description for hashing",
		"agentType": "command",
		"categories": ["AI"],
		"capabilities": [{"name": "cap"}],
		"commands": [{
			"trigger": "small",
			"description": "Uses legacy choice metadata",
			"pricePerUnit": 0.1,
			"priceType": "task-transaction",
			"taskUnit": "per-query",
			"parameters": [{
				"name": "template_title",
				"type": "choice",
				"required": true,
				"description": "Choose a template",
				"choices": ["Trade Tracking Vault", "Daily Journal"]
			}]
		}]
	}`)

	var legacyConfig AgentConfig
	if err := json.Unmarshal(legacyJSON, &legacyConfig); err != nil {
		t.Fatalf("failed to unmarshal legacy config: %v", err)
	}

	canonicalConfig := &AgentConfig{
		Name:         "Test Agent",
		AgentID:      "test-agent",
		Description:  "A valid description for hashing",
		AgentType:    "command",
		Categories:   []string{"AI"},
		Capabilities: []Capability{{Name: "cap"}},
		Commands: []Command{{
			Trigger:      "small",
			Description:  "Uses legacy choice metadata",
			PricePerUnit: 0.1,
			PriceType:    "task-transaction",
			TaskUnit:     "per-query",
			Parameters: []CommandParameter{{
				Name:        "template_title",
				Type:        "enum",
				Required:    true,
				Description: "Choose a template",
				Options:     []string{"Trade Tracking Vault", "Daily Journal"},
			}},
		}},
	}

	if got, want := legacyConfig.Commands[0].Parameters[0].Type, ParamTypeEnum; got != want {
		t.Fatalf("legacy parameter type = %q, want %q", got, want)
	}
	if len(legacyConfig.Commands[0].Parameters[0].Options) != 2 {
		t.Fatalf("legacy parameter options were not normalized: %#v", legacyConfig.Commands[0].Parameters[0].Options)
	}

	legacyHash := GenerateConfigHash(&legacyConfig)
	canonicalHash := GenerateConfigHash(canonicalConfig)
	if legacyHash != canonicalHash {
		t.Fatalf("normalized legacy hash %q did not match canonical hash %q", legacyHash, canonicalHash)
	}
}

func TestAgentConfigUnmarshalSupportsSnakeCaseAndLegacyCamelCase(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "snake_case",
			raw: `{
				"name": "Test Agent",
				"agent_id": "test-agent",
				"short_description": "Short summary",
				"description": "A valid description for parsing",
				"agent_type": "command",
				"nlp_fallback": true,
				"categories": ["AI"],
				"capabilities": [{"name": "cap"}],
				"commands": []
			}`,
		},
		{
			name: "legacy camelCase",
			raw: `{
				"name": "Test Agent",
				"agentId": "test-agent",
				"shortDescription": "Short summary",
				"description": "A valid description for parsing",
				"agentType": "command",
				"nlpFallback": true,
				"categories": ["AI"],
				"capabilities": [{"name": "cap"}],
				"commands": []
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config AgentConfig
			if err := json.Unmarshal([]byte(tt.raw), &config); err != nil {
				t.Fatalf("failed to unmarshal config: %v", err)
			}

			if config.AgentID != "test-agent" {
				t.Fatalf("AgentID = %q, want %q", config.AgentID, "test-agent")
			}
			if config.ShortDescription != "Short summary" {
				t.Fatalf("ShortDescription = %q, want %q", config.ShortDescription, "Short summary")
			}
			if config.AgentType != "command" {
				t.Fatalf("AgentType = %q, want %q", config.AgentType, "command")
			}
			if !config.NlpFallback {
				t.Fatal("NlpFallback = false, want true")
			}
		})
	}
}

func TestAgentConfigUnmarshalPrefersCanonicalSnakeCaseWhenBothPresent(t *testing.T) {
	raw := []byte(`{
		"name": "Test Agent",
		"agent_id": "snake-agent",
		"agentId": "camel-agent",
		"short_description": "Snake summary",
		"shortDescription": "Camel summary",
		"description": "A valid description for parsing",
		"agent_type": "commandless",
		"agentType": "command",
		"nlp_fallback": true,
		"nlpFallback": false,
		"mcp_manifest": "https://snake.example/manifest.json",
		"mcpManifest": "https://camel.example/manifest.json",
		"tutorial_url": "https://snake.example/tutorial",
		"tutorialUrl": "https://camel.example/tutorial",
		"faq_items": [{"question": "snake", "answer": "wins"}],
		"faqItems": [{"question": "camel", "answer": "loses"}],
		"categories": ["AI"],
		"capabilities": [{"name": "cap"}],
		"commands": []
	}`)

	var config AgentConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if config.AgentID != "snake-agent" {
		t.Fatalf("AgentID = %q, want %q", config.AgentID, "snake-agent")
	}
	if config.ShortDescription != "Snake summary" {
		t.Fatalf("ShortDescription = %q, want %q", config.ShortDescription, "Snake summary")
	}
	if config.AgentType != "commandless" {
		t.Fatalf("AgentType = %q, want %q", config.AgentType, "commandless")
	}
	if !config.NlpFallback {
		t.Fatal("NlpFallback = false, want true")
	}
	if config.McpManifest != "https://snake.example/manifest.json" {
		t.Fatalf("McpManifest = %q, want snake value", config.McpManifest)
	}
	if config.TutorialURL != "https://snake.example/tutorial" {
		t.Fatalf("TutorialURL = %q, want snake value", config.TutorialURL)
	}
	if len(config.FAQItems) != 1 || config.FAQItems[0].Question != "snake" {
		t.Fatalf("FAQItems = %#v, want canonical snake_case items", config.FAQItems)
	}
}

func TestPreValidate(t *testing.T) {
	minter := &Minter{}

	tests := []struct {
		name    string
		config  *AgentConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &AgentConfig{
				Name:      "Test Agent",
				AgentID:   "test-agent",
				AgentType: "command",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			config: &AgentConfig{
				AgentID:   "test-agent",
				AgentType: "command",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing agentId",
			config: &AgentConfig{
				Name:      "Test",
				AgentType: "command",
			},
			wantErr: true,
			errMsg:  "agent_id is required",
		},
		{
			name: "missing agentType",
			config: &AgentConfig{
				Name:    "Test",
				AgentID: "test",
			},
			wantErr: true,
			errMsg:  "agent_type is required",
		},
		{
			name: "invalid agentId with uppercase",
			config: &AgentConfig{
				Name:      "Test",
				AgentID:   "Test-Agent",
				AgentType: "command",
			},
			wantErr: true,
			errMsg:  "lowercase",
		},
		{
			name: "invalid agentId with space",
			config: &AgentConfig{
				Name:      "Test",
				AgentID:   "test agent",
				AgentType: "command",
			},
			wantErr: true,
			errMsg:  "lowercase",
		},
		{
			name: "valid agentId with hyphen and numbers",
			config: &AgentConfig{
				Name:      "Test",
				AgentID:   "my-agent-123",
				AgentType: "command",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := minter.preValidate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("preValidate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("preValidate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	minter := &Minter{}

	tests := []struct {
		name    string
		config  *AgentConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid full config",
			config: &AgentConfig{
				Name:        "Test Agent",
				AgentID:     "test-agent",
				Description: "This is a test agent with enough description",
				AgentType:   "command",
				Categories:  []string{"AI"},
				Capabilities: []Capability{
					{Name: "capability1", Description: "does something"},
				},
			},
			wantErr: false,
		},
		{
			name: "name too short",
			config: &AgentConfig{
				Name:         "Ab",
				AgentID:      "test",
				Description:  "Valid description here",
				AgentType:    "command",
				Categories:   []string{"AI"},
				Capabilities: []Capability{{Name: "cap"}},
			},
			wantErr: true,
			errMsg:  "at least 3",
		},
		{
			name: "description too short",
			config: &AgentConfig{
				Name:         "Valid Name",
				AgentID:      "test",
				Description:  "Short",
				AgentType:    "command",
				Categories:   []string{"AI"},
				Capabilities: []Capability{{Name: "cap"}},
			},
			wantErr: true,
			errMsg:  "at least 10",
		},
		{
			name: "invalid agentType",
			config: &AgentConfig{
				Name:         "Valid Name",
				AgentID:      "test",
				Description:  "Valid description here",
				AgentType:    "invalid",
				Categories:   []string{"AI"},
				Capabilities: []Capability{{Name: "cap"}},
			},
			wantErr: true,
			errMsg:  "command",
		},
		{
			name: "no categories",
			config: &AgentConfig{
				Name:         "Valid Name",
				AgentID:      "test",
				Description:  "Valid description here",
				AgentType:    "command",
				Categories:   []string{},
				Capabilities: []Capability{{Name: "cap"}},
			},
			wantErr: true,
			errMsg:  "category",
		},
		{
			name: "too many categories",
			config: &AgentConfig{
				Name:         "Valid Name",
				AgentID:      "test",
				Description:  "Valid description here",
				AgentType:    "command",
				Categories:   []string{"AI", "Automation", "Finance"},
				Capabilities: []Capability{{Name: "cap"}},
			},
			wantErr: true,
			errMsg:  "2",
		},
		{
			name: "no capabilities",
			config: &AgentConfig{
				Name:         "Valid Name",
				AgentID:      "test",
				Description:  "Valid description here",
				AgentType:    "command",
				Categories:   []string{"AI"},
				Capabilities: []Capability{},
			},
			wantErr: true,
			errMsg:  "capability",
		},
		{
			name: "mcp type without manifest",
			config: &AgentConfig{
				Name:         "Valid Name",
				AgentID:      "test",
				Description:  "Valid description here",
				AgentType:    "mcp",
				Categories:   []string{"AI"},
				Capabilities: []Capability{{Name: "cap"}},
				McpManifest:  "",
			},
			wantErr: true,
			errMsg:  "mcpManifest",
		},
		{
			name: "mcp type with manifest",
			config: &AgentConfig{
				Name:         "Valid Name",
				AgentID:      "test",
				Description:  "Valid description here",
				AgentType:    "mcp",
				Categories:   []string{"AI"},
				Capabilities: []Capability{{Name: "cap"}},
				McpManifest:  "https://example.com/manifest.json",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := minter.validateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("validateConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func boolPtr(b bool) *bool { return &b }
func intPtr(i int) *int    { return &i }

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestFileSizeLimit(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Test file too large
	largePath := filepath.Join(tmpDir, "large.json")
	largeContent := make([]byte, DefaultMaxJSONSize+1)
	for i := range largeContent {
		largeContent[i] = 'x'
	}
	if err := os.WriteFile(largePath, largeContent, 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	minter := &Minter{
		httpClient: NewHTTPClient("http://localhost:8080"),
		walClient:  NewWALClient(),
	}

	_, err := minter.Mint(largePath)
	if err == nil {
		t.Error("Expected error for large file, got nil")
	}
	if !contains(err.Error(), "too large") {
		t.Errorf("Expected 'too large' error, got: %v", err)
	}
}

func TestValidateConfigAcceptsLegacyChoiceParameterJSON(t *testing.T) {
	raw := []byte(`{
		"name": "Choice Agent",
		"agentId": "choice-agent",
		"description": "A valid description that is long enough",
		"agentType": "command",
		"categories": ["AI"],
		"capabilities": [{"name": "cap"}],
		"commands": [{
			"trigger": "medium",
			"description": "Legacy parameter metadata",
			"pricePerUnit": 0.1,
			"priceType": "task-transaction",
			"taskUnit": "per-query",
			"parameters": [{
				"name": "template_title",
				"type": "choice",
				"required": true,
				"enum": ["Trade Tracking Vault", "Daily Journal"]
			}]
		}]
	}`)

	var config AgentConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	minter := &Minter{}
	if err := minter.validateConfig(&config); err != nil {
		t.Fatalf("validateConfig rejected legacy choice metadata: %v", err)
	}

	param := config.Commands[0].Parameters[0]
	if param.Type != ParamTypeEnum {
		t.Fatalf("normalized parameter type = %q, want %q", param.Type, ParamTypeEnum)
	}
	if len(param.Options) != 2 {
		t.Fatalf("normalized parameter options = %#v, want two entries", param.Options)
	}
}
