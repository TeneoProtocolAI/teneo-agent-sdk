package main

import (
	"fmt"
	"log"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/naming"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
)

func main() {
	fmt.Println("🏷️  Agent Naming Conventions Example")
	fmt.Println("===================================")

	// demonstrate basic validation
	demonstrateBasicValidation()

	// demonstrate name normalization
	demonstrateNormalization()

	// demonstrate name generation
	demonstrateGeneration()

	// demonstrate custom rules
	demonstrateCustomRules()

	// demonstrate suggestions
	demonstrateSuggestions()

	// demonstrate agent config integration
	demonstrateAgentConfigIntegration()
	demonstrateReservedNames()
}

func demonstrateBasicValidation() {
	fmt.Println("\n📋 Basic Name Validation")
	fmt.Println("========================")

	validator := naming.NewDefaultValidator()

	testNames := []string{
		"security-scanner",  // valid
		"data-processor-v2", // valid
		"API_Gateway",       // will be normalized
		"123invalid",        // invalid - starts with number
		"agent@name!",       // invalid - special characters
		"system",            // invalid - reserved name
		"a",                 // invalid - too short
		"this-is-a-very-long-agent-name-that-exceeds-maximum-length", // invalid - too long
	}

	for _, name := range testNames {
		result := validator.ValidateName(name)

		fmt.Printf("Name: %-50s ", fmt.Sprintf("'%s'", name))
		if result.IsValid {
			fmt.Printf("✅ VALID   -> %s\n", result.NormalizedName)
		} else {
			fmt.Printf("❌ INVALID -> %v\n", result.Errors[0])
		}

		if len(result.Warnings) > 0 {
			fmt.Printf("     ⚠️  Warning: %s\n", result.Warnings[0])
		}
	}
}

func demonstrateNormalization() {
	fmt.Println("\n🔧 Name Normalization")
	fmt.Println("====================")

	validator := naming.NewDefaultValidator()

	testCases := []string{
		"Security Scanner",   // spaces
		"DATA_PROCESSOR",     // uppercase + underscores
		"agent@name!",        // invalid characters
		"MyAwesome-Agent",    // mixed case
		"api_gateway_v2",     // underscores
		"123StartWithNumber", // starts with number
	}

	for _, original := range testCases {
		normalized := validator.NormalizeName(original)
		fmt.Printf("Original:   %-25s -> Normalized: %s\n", fmt.Sprintf("'%s'", original), normalized)
	}
}

func demonstrateGeneration() {
	fmt.Println("\n🎯 Name Generation")
	fmt.Println("=================")

	validator := naming.NewDefaultValidator()

	testCases := []struct {
		baseName string
		purpose  string
	}{
		{"security", "scanner"},
		{"data", "processor"},
		{"", "monitor"}, // empty base name
		{"api", ""},     // empty purpose
		{"", ""},        // both empty
		{"blockchain", "validator"},
		{"ml", "inference"},
	}

	for _, tc := range testCases {
		generated := validator.GenerateName(tc.baseName, tc.purpose)
		fmt.Printf("Base: %-10s Purpose: %-10s -> Generated: %s\n",
			fmt.Sprintf("'%s'", tc.baseName),
			fmt.Sprintf("'%s'", tc.purpose),
			generated)
	}
}

func demonstrateCustomRules() {
	fmt.Println("\n⚙️  Custom Naming Rules")
	fmt.Println("======================")

	// use strict validator for production
	strictValidator := naming.NewStrictValidator()

	testNames := []string{
		"security-scanner-agent",       // valid for strict
		"data-proc",                    // too short
		"SecurityAgent",                // invalid case + missing suffix
		"api_gateway_agent",            // underscores not allowed
		"fraud-detection-system-agent", // too long
	}

	fmt.Println("Using Strict Production Rules:")
	fmt.Println("- Min length: 5 chars")
	fmt.Println("- Max length: 30 chars")
	fmt.Println("- Must be lowercase")
	fmt.Println("- Must end with '-agent'")
	fmt.Println("- No underscores allowed")
	fmt.Println()

	for _, name := range testNames {
		result := strictValidator.ValidateName(name)

		fmt.Printf("Name: %-30s ", fmt.Sprintf("'%s'", name))
		if result.IsValid {
			fmt.Printf("✅ VALID\n")
		} else {
			fmt.Printf("❌ INVALID -> %s\n", result.Errors[0])
		}
	}
}

func demonstrateSuggestions() {
	fmt.Println("\n💡 Name Suggestions")
	fmt.Println("==================")

	validator := naming.NewDefaultValidator()

	invalidNames := []string{
		"123InvalidAgent",
		"agent@name!",
		"system",
		"VeryLongAgentNameThatExceedsMaximumLengthAllowed",
		"a",
	}

	for _, invalidName := range invalidNames {
		fmt.Printf("Invalid name: '%s'\n", invalidName)

		suggestions := validator.SuggestNames(invalidName, 3)
		if len(suggestions) > 0 {
			fmt.Printf("Suggestions: %v\n", suggestions)
		} else {
			fmt.Printf("No valid suggestions could be generated\n")
		}
		fmt.Println()
	}
}

func demonstrateAgentConfigIntegration() {
	fmt.Println("\n🔧 Agent Config Integration")
	fmt.Println("===========================")

	validator := naming.NewDefaultValidator()

	// example configurations
	configs := []*types.AgentConfig{
		{
			Name: "security-scanner-v2",
			NamingRules: &types.AgentNamingRules{
				MaxLength:        30,
				MinLength:        5,
				CaseSensitive:    false,
				AllowNumbers:     true,
				AllowHyphens:     true,
				AllowUnderscores: false,
			},
		},
		{
			Name: "InvalidName123!",
			NamingRules: &types.AgentNamingRules{
				MaxLength:        20,
				MinLength:        8,
				CaseSensitive:    true,
				AllowNumbers:     false,
				AllowHyphens:     true,
				AllowUnderscores: false,
				RequiredSuffix:   "-agent",
			},
		},
		{
			Name: "system", // reserved name
		},
	}

	for i, config := range configs {
		fmt.Printf("Config %d: Agent name '%s'\n", i+1, config.Name)

		result := validator.ValidateAgentConfig(config)

		if result.IsValid {
			fmt.Printf("  ✅ Valid agent configuration\n")
			if result.NormalizedName != config.Name {
				fmt.Printf("  📝 Normalized: %s\n", result.NormalizedName)
			}
		} else {
			fmt.Printf("  ❌ Invalid agent configuration\n")
			for _, err := range result.Errors {
				fmt.Printf("     - %s\n", err)
			}
		}

		if len(result.Warnings) > 0 {
			fmt.Printf("  ⚠️  Warnings:\n")
			for _, warning := range result.Warnings {
				fmt.Printf("     - %s\n", warning)
			}
		}

		fmt.Println()
	}
}

// demonstrateReservedNames shows all reserved names
func demonstrateReservedNames() {
	fmt.Println("\n🚫 Reserved Names")
	fmt.Println("================")

	validator := naming.NewDefaultValidator()
	rules := validator.GetRules()

	fmt.Println("The following names are reserved and cannot be used:")

	categories := map[string][]string{
		"System":      {"system", "admin", "root", "coordinator", "manager"},
		"Protocol":    {"teneo", "protocol", "network", "blockchain", "validator"},
		"Service":     {"api", "gateway", "proxy", "health", "metrics"},
		"Common":      {"agent", "bot", "service", "handler", "processor"},
		"Development": {"test", "demo", "example", "mock", "debug"},
	}

	for category, names := range categories {
		fmt.Printf("\n%s Reserved:\n", category)
		for _, name := range names {
			if rules.ReservedNames[name] {
				fmt.Printf("  - %s\n", name)
			}
		}
	}
}

func init() {
	// demonstrate reserved names at startup
	go func() {
		// This could be shown in help or documentation
		log.Printf("Agent Naming Conventions loaded with %d reserved names",
			len(naming.DefaultAgentNamingRules.ReservedNames))
	}()
}
