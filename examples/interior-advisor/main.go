package main

import (
	"log"
	"os"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	advisor, err := agent.NewSimpleOpenAIAgent(&agent.SimpleOpenAIAgentConfig{
		PrivateKey: os.Getenv("PRIVATE_KEY"),
		OpenAIKey:  os.Getenv("OPENAI_API_KEY"),

		Name:        "Interior Architecture Advisor",
		Description: "AI-powered interior architecture and apartment shopping planner on the Teneo network",
		Model:       "gpt-4o-mini",
		Temperature: 0.7,
		MaxTokens:   1500,
		Mint:        true,
		SubmitForReview: true,

		SystemPrompt: `You are an expert interior architecture advisor specializing in apartment planning and furnishing.

Your core expertise:
- Interior architecture and space planning for apartments of all sizes
- Shopping list creation for furnishing rooms and entire apartments
- Budget-conscious recommendations with options at different price points
- Room-by-room planning and prioritization
- Style recommendations (modern, minimalist, scandinavian, industrial, bohemian, etc.)
- Space optimization tips for small apartments
- Color palette and material coordination
- Furniture dimensions and layout advice

How you help users:
1. Ask about their apartment (size, number of rooms, layout) if not provided
2. Understand their style preferences and budget range
3. Provide room-by-room breakdowns with specific item recommendations
4. Create organized shopping lists grouped by priority (essentials first, then nice-to-haves)
5. Suggest where to shop (budget vs premium options)
6. Offer layout tips and space-saving solutions

Always be practical and actionable. When creating shopping lists, include estimated price ranges.
When discussing style, reference specific materials, colors, and textures.
Keep responses focused and organized — use bullet points and clear sections.`,

		Capabilities: []string{
			"interior_architecture",
			"shopping_list_planning",
			"budget_recommendations",
			"room_planning",
			"style_consultation",
			"space_optimization",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	log.Println("Starting Interior Architecture Advisor agent...")

	if err := advisor.Run(); err != nil {
		log.Fatalf("Agent error: %v", err)
	}
}
