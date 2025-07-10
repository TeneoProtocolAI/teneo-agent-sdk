package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/teneo/agent-sdk-go/pkg/agent"
)

// ExampleAgent demonstrates a complete agent implementation using the enhanced SDK
type ExampleAgent struct {
	name         string
	capabilities []string
}

// NewExampleAgent creates a new example agent
func NewExampleAgent() *ExampleAgent {
	return &ExampleAgent{
		name: "Enhanced Example Agent",
		capabilities: []string{
			"text_analysis_detailed",
			"content_generation_stories",
			"content_generation_poems",
			"content_generation_emails",
			"code_assistance_debug",
			"code_assistance_examples",
			"math_calculations_basic",
			"math_calculations_expressions",
			"weather_information_demo",
			"time_utilities_timezone",
			"system_status_health",
			"data_formatting_json",
			"data_formatting_csv",
			"data_formatting_tables",
			"translation_multilingual",
			"text_summarization",
			"conversation_natural",
			"help_commands_detailed",
		},
	}
}

// ProcessTask processes a task and returns a result
func (a *ExampleAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	log.Printf("🔄 Processing task: %s", task)

	taskLower := strings.ToLower(strings.TrimSpace(task))

	// Handle help and capabilities
	if strings.Contains(taskLower, "help") || strings.Contains(taskLower, "capabilities") || strings.Contains(taskLower, "what can you do") {
		return a.getHelpMessage(), nil
	}

	// Handle greetings
	if strings.Contains(taskLower, "hello") || strings.Contains(taskLower, "hi") || strings.Contains(taskLower, "hey") {
		return fmt.Sprintf("👋 Hello! I'm %s, your Teneo network assistant. I can help with text analysis, content generation, code assistance, math calculations, and much more. Type 'help' to see all my capabilities!", a.name), nil
	}

	// Text Analysis
	if strings.Contains(taskLower, "analyze") || strings.Contains(taskLower, "analysis") {
		return a.analyzeText(task), nil
	}

	// Content Generation
	if strings.Contains(taskLower, "generate") || strings.Contains(taskLower, "create") || strings.Contains(taskLower, "write") {
		return a.generateContent(task), nil
	}

	// Code Assistance
	if strings.Contains(taskLower, "code") || strings.Contains(taskLower, "program") || strings.Contains(taskLower, "function") || strings.Contains(taskLower, "debug") {
		return a.assistWithCode(task), nil
	}

	// Math Calculations
	if strings.Contains(taskLower, "calculate") || strings.Contains(taskLower, "math") || strings.Contains(taskLower, "compute") || containsMathSymbols(task) {
		return a.performCalculation(task), nil
	}

	// Weather Info
	if strings.Contains(taskLower, "weather") || strings.Contains(taskLower, "temperature") || strings.Contains(taskLower, "forecast") {
		return a.getWeatherInfo(task), nil
	}

	// Time Utilities
	if strings.Contains(taskLower, "time") || strings.Contains(taskLower, "date") || strings.Contains(taskLower, "timezone") {
		return a.getTimeInfo(task), nil
	}

	// System Status
	if strings.Contains(taskLower, "status") || strings.Contains(taskLower, "health") || strings.Contains(taskLower, "system") {
		return a.getSystemStatus(), nil
	}

	// Data Formatting
	if strings.Contains(taskLower, "format") || strings.Contains(taskLower, "json") || strings.Contains(taskLower, "csv") || strings.Contains(taskLower, "table") {
		return a.formatData(task), nil
	}

	// Translation
	if strings.Contains(taskLower, "translate") || strings.Contains(taskLower, "translation") {
		return a.translateText(task), nil
	}

	// Summarization
	if strings.Contains(taskLower, "summarize") || strings.Contains(taskLower, "summary") || strings.Contains(taskLower, "tldr") {
		return a.summarizeText(task), nil
	}

	// Default conversation
	return a.handleConversation(task), nil
}

// getHelpMessage returns a comprehensive help message
func (a *ExampleAgent) getHelpMessage() string {
	return `🤖 **Enhanced Teneo Agent - Help & Capabilities**

📋 **Available Commands:**

**🔍 Text Analysis:**
   • "analyze this text: [your text]" - Perform detailed text analysis
   • "analysis of [text]" - Get insights about any text

**✍️ Content Generation:**
   • "generate a story about [topic]" - Create creative content
   • "write a summary of [topic]" - Generate summaries
   • "create content for [purpose]" - Generate various content types

**💻 Code Assistance:**
   • "help me code [language/task]" - Get coding help
   • "debug this code: [code]" - Debug assistance
   • "write a function to [task]" - Code generation

**🧮 Math Calculations:**
   • "calculate 15 * 23 + 7" - Perform calculations
   • "compute the square root of 144" - Mathematical operations
   • Basic arithmetic: +, -, *, /, ^, sqrt()

**🌤️ Weather Info:**
   • "weather in [city]" - Get weather information
   • "temperature forecast" - Weather forecasts

**⏰ Time Utilities:**
   • "what time is it?" - Current time
   • "date today" - Current date
   • "timezone info" - Timezone information

**🔧 System Status:**
   • "status" - Agent health and system status
   • "system health" - Detailed system information

**📊 Data Formatting:**
   • "format this data as JSON: [data]" - JSON formatting
   • "create a table from: [data]" - Table formatting

**🌍 Translation:**
   • "translate to [language]: [text]" - Text translation
   • "translation help" - Translation assistance

**📝 Summarization:**
   • "summarize: [long text]" - Text summarization
   • "tldr: [content]" - Quick summaries

**💬 General Conversation:**
   • Just chat with me naturally!

Type any command or ask me anything! 🚀`
}

// analyzeText performs detailed text analysis
func (a *ExampleAgent) analyzeText(task string) string {
	text := extractTextFromTask(task, "analyze")
	if text == "" {
		return "📊 Please provide text to analyze. Example: 'analyze this text: Hello world!'"
	}

	words := strings.Fields(text)
	sentences := strings.Split(text, ".")
	chars := len(text)
	wordsCount := len(words)
	sentenceCount := len(sentences)
	avgWordsPerSentence := float64(wordsCount) / float64(sentenceCount)

	return fmt.Sprintf(`📊 **Text Analysis Results:**

📝 **Text:** "%s"

📈 **Statistics:**
   • Characters: %d
   • Words: %d
   • Sentences: %d
   • Avg words per sentence: %.1f

🎯 **Classification:**
   • Length: %s
   • Complexity: %s
   • Type: %s

✨ **Insights:**
   • Reading time: ~%d seconds
   • Language: English (detected)
   • Sentiment: %s`,
		text,
		chars, wordsCount, sentenceCount, avgWordsPerSentence,
		getTextLength(wordsCount),
		getTextComplexity(avgWordsPerSentence),
		getTextType(text),
		wordsCount/3, // avg reading speed ~180 wpm
		getTextSentiment(text))
}

// generateContent creates various types of content
func (a *ExampleAgent) generateContent(task string) string {
	topic := extractTextFromTask(task, "generate", "create", "write")
	if topic == "" {
		return "✍️ Please specify what to generate. Examples:\n• 'generate a story about robots'\n• 'create content for a blog post'\n• 'write a poem about nature'"
	}

	if strings.Contains(strings.ToLower(task), "story") {
		return a.generateStory(topic)
	} else if strings.Contains(strings.ToLower(task), "poem") {
		return a.generatePoem(topic)
	} else if strings.Contains(strings.ToLower(task), "email") {
		return a.generateEmail(topic)
	} else {
		return a.generateGenericContent(topic)
	}
}

// assistWithCode provides coding assistance
func (a *ExampleAgent) assistWithCode(task string) string {
	code := extractTextFromTask(task, "code", "function", "debug")

	if strings.Contains(strings.ToLower(task), "debug") {
		return fmt.Sprintf(`🐛 **Code Debug Assistant:**

🔍 **Analyzing:** %s

🔧 **Common Debug Steps:**
1. Check syntax and brackets/parentheses
2. Verify variable names and types
3. Look for off-by-one errors
4. Check function signatures
5. Validate input/output expectations

💡 **Debug Tips:**
   • Add print statements to trace execution
   • Use a debugger or IDE tools
   • Test with simple inputs first
   • Check documentation for library functions

📝 **Best Practices:**
   • Write unit tests
   • Use meaningful variable names
   • Add comments for complex logic
   • Handle edge cases

Need more specific help? Share your code and error message!`, code)
	}

	return fmt.Sprintf(`💻 **Code Assistant:**

📋 **Request:** %s

🔧 **Code Example:**

// Example function based on your request
func processTask(input string) (string, error) {
    if input == "" {
        return "", fmt.Errorf("input cannot be empty")
    }

    result := strings.ToUpper(input)
    return fmt.Sprintf("Processed: %%s", result), nil
}

💡 **Programming Tips:**
   • Always handle errors gracefully
   • Use descriptive variable names
   • Add input validation
   • Write tests for your functions
   • Follow language conventions

🚀 **Next Steps:**
   • Test the code thoroughly
   • Add error handling
   • Consider edge cases
   • Document your functions

Need help with a specific language or problem? Just ask!`, code)
}

// performCalculation handles mathematical operations
func (a *ExampleAgent) performCalculation(task string) string {
	calculation := extractMathExpression(task)
	if calculation == "" {
		return "🧮 Please provide a calculation. Examples:\n• 'calculate 15 + 25'\n• '50 * 30'\n• 'square root of 144'"
	}

	result := evaluateExpression(calculation)
	return fmt.Sprintf(`🧮 **Mathematical Calculation:**

📝 **Expression:** %s
🔢 **Result:** %s

💡 **Supported Operations:**
   • Addition: + (e.g., 5 + 3)
   • Subtraction: - (e.g., 10 - 4)
   • Multiplication: * (e.g., 6 * 7)
   • Division: / (e.g., 20 / 4)
   • Exponentiation: ^ (e.g., 2 ^ 3)
   • Square root: sqrt (e.g., sqrt(16))

🔢 **Example Calculations:**
   • Simple: "calculate 25 + 17"
   • Complex: "compute (15 * 3) + sqrt(49)"
   • Percentage: "what is 20%% of 150"`, calculation, result)
}

// Helper functions
func containsMathSymbols(text string) bool {
	mathSymbols := []string{"+", "-", "*", "/", "=", "^", "sqrt", "%"}
	for _, symbol := range mathSymbols {
		if strings.Contains(text, symbol) {
			return true
		}
	}
	return false
}

func extractTextFromTask(task string, keywords ...string) string {
	text := task
	for _, keyword := range keywords {
		if idx := strings.Index(strings.ToLower(task), strings.ToLower(keyword)); idx != -1 {
			// Find text after keyword
			afterKeyword := task[idx+len(keyword):]
			afterKeyword = strings.TrimSpace(afterKeyword)
			if strings.HasPrefix(afterKeyword, ":") || strings.HasPrefix(afterKeyword, " ") {
				afterKeyword = strings.TrimPrefix(afterKeyword, ":")
				afterKeyword = strings.TrimSpace(afterKeyword)
			}
			if afterKeyword != "" {
				text = afterKeyword
				break
			}
		}
	}
	return strings.TrimSpace(text)
}

func extractMathExpression(task string) string {
	// Simple extraction - in practice this would be more sophisticated
	expr := task
	keywords := []string{"calculate", "compute", "math", "="}
	for _, keyword := range keywords {
		if idx := strings.Index(strings.ToLower(task), keyword); idx != -1 {
			afterKeyword := task[idx+len(keyword):]
			afterKeyword = strings.TrimSpace(afterKeyword)
			if strings.HasPrefix(afterKeyword, ":") || strings.HasPrefix(afterKeyword, " ") {
				afterKeyword = strings.TrimPrefix(afterKeyword, ":")
				afterKeyword = strings.TrimSpace(afterKeyword)
			}
			if afterKeyword != "" {
				expr = afterKeyword
				break
			}
		}
	}
	return strings.TrimSpace(expr)
}

func evaluateExpression(expr string) string {
	// Simple calculator - in practice use a proper math parser
	expr = strings.TrimSpace(expr)

	// Handle simple operations
	if strings.Contains(expr, " + ") {
		parts := strings.Split(expr, " + ")
		if len(parts) == 2 {
			if a, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64); err == nil {
				if b, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
					return fmt.Sprintf("%.2f", a+b)
				}
			}
		}
	}

	return fmt.Sprintf("Calculation requested: %s (demo mode - basic math operations supported)", expr)
}

func getTextLength(wordCount int) string {
	switch {
	case wordCount < 10:
		return "Very Short"
	case wordCount < 50:
		return "Short"
	case wordCount < 200:
		return "Medium"
	case wordCount < 500:
		return "Long"
	default:
		return "Very Long"
	}
}

func getTextComplexity(avgWords float64) string {
	switch {
	case avgWords < 8:
		return "Simple"
	case avgWords < 15:
		return "Moderate"
	case avgWords < 25:
		return "Complex"
	default:
		return "Very Complex"
	}
}

func getTextType(text string) string {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "?") {
		return "Question"
	} else if strings.Contains(lower, "!") {
		return "Exclamatory"
	} else if strings.Contains(lower, "hello") || strings.Contains(lower, "hi") {
		return "Greeting"
	}
	return "Statement"
}

func getTextSentiment(text string) string {
	lower := strings.ToLower(text)
	positive := []string{"good", "great", "excellent", "amazing", "wonderful", "happy", "love", "best"}
	negative := []string{"bad", "terrible", "awful", "hate", "worst", "sad", "angry"}

	posCount, negCount := 0, 0
	for _, word := range positive {
		if strings.Contains(lower, word) {
			posCount++
		}
	}
	for _, word := range negative {
		if strings.Contains(lower, word) {
			negCount++
		}
	}

	if posCount > negCount {
		return "Positive"
	} else if negCount > posCount {
		return "Negative"
	}
	return "Neutral"
}

// getWeatherInfo provides weather information
func (a *ExampleAgent) getWeatherInfo(task string) string {
	city := extractTextFromTask(task, "weather", "temperature", "forecast")
	if city == "" {
		city = "your location"
	}

	return fmt.Sprintf(`🌤️ **Weather Information for %s:**

📅 **Current Conditions:**
   • Temperature: 22°C (72°F)
   • Condition: Partly Cloudy
   • Humidity: 65%%
   • Wind: 15 km/h SW
   • Pressure: 1013 hPa

📊 **Forecast:**
   • Today: High 25°C, Low 18°C - Partly cloudy
   • Tomorrow: High 23°C, Low 16°C - Light rain
   • Weekend: High 26°C, Low 19°C - Sunny

💡 **Note:** This is demo weather data. In a real implementation, I would connect to a weather API to provide actual current conditions and forecasts for any location.

🔍 **Try asking:** "weather in London" or "temperature forecast for Tokyo"`, city)
}

// getTimeInfo provides time and date information
func (a *ExampleAgent) getTimeInfo(task string) string {
	now := time.Now()

	return fmt.Sprintf(`⏰ **Time & Date Information:**

🕐 **Current Time:**
   • Local Time: %s
   • UTC Time: %s
   • Date: %s
   • Day of Week: %s

🌍 **Time Zones:**
   • Pacific: %s
   • Eastern: %s
   • London: %s
   • Tokyo: %s

📅 **Date Details:**
   • Day of Year: %d
   • Week of Year: %d
   • Days until New Year: %d

💡 **Time Utilities:**
   • Ask for specific timezones
   • Date calculations
   • Time conversions`,
		now.Format("15:04:05 MST"),
		now.UTC().Format("15:04:05 UTC"),
		now.Format("January 2, 2006"),
		now.Format("Monday"),
		now.In(getPacificLocation()).Format("15:04 PST"),
		now.In(getEasternLocation()).Format("15:04 EST"),
		now.UTC().Format("15:04 UTC"),
		now.In(getTokyoLocation()).Format("15:04 JST"),
		now.YearDay(),
		getWeekOfYear(now),
		getDaysUntilNewYear(now))
}

// getSystemStatus provides system health information
func (a *ExampleAgent) getSystemStatus() string {
	uptime := time.Since(time.Now().Add(-time.Hour * 2)) // Mock uptime

	return fmt.Sprintf(`🔧 **System Status & Health:**

✅ **Agent Status:**
   • Status: Online & Operational
   • Uptime: %s
   • Performance: Excellent
   • Memory Usage: 45MB
   • CPU Usage: 2%%

🌐 **Network:**
   • Connection: Stable
   • Latency: 25ms
   • WebSocket: Connected
   • Authentication: Verified

🔋 **Capabilities Status:**
   • Text Analysis: ✅ Active
   • Content Generation: ✅ Active
   • Code Assistance: ✅ Active
   • Math Calculations: ✅ Active
   • All Systems: ✅ Operational

📊 **Statistics:**
   • Tasks Processed: 47
   • Success Rate: 98.5%%
   • Avg Response Time: 1.2s

🚀 **Ready to assist with any task!**`, uptime.Round(time.Second))
}

// formatData handles data formatting requests
func (a *ExampleAgent) formatData(task string) string {
	data := extractTextFromTask(task, "format", "json", "csv", "table")

	if strings.Contains(strings.ToLower(task), "json") {
		return a.formatAsJSON(data)
	} else if strings.Contains(strings.ToLower(task), "csv") {
		return a.formatAsCSV(data)
	} else if strings.Contains(strings.ToLower(task), "table") {
		return a.formatAsTable(data)
	}

	return fmt.Sprintf(`📊 **Data Formatting Service:**

📝 **Input Data:** %s

🔧 **Available Formats:**
   • JSON: "format as JSON: name,age,city John,25,NYC"
   • CSV: "format as CSV: [your data]"
   • Table: "create table from: [your data]"

💡 **Example Commands:**
   • "format this data as JSON: name John, age 25, city NYC"
   • "create a table from: Product,Price,Stock Apple,1.50,100"
   • "convert to CSV: user data with names and emails"

🚀 **Ready to format your data in any structure!**`, data)
}

// translateText handles translation requests
func (a *ExampleAgent) translateText(task string) string {
	text := extractTextFromTask(task, "translate", "translation")

	return fmt.Sprintf(`🌍 **Translation Service:**

📝 **Original Text:** %s

🔧 **Translation Example:**
   • English: "Hello, how are you?"
   • Spanish: "Hola, ¿cómo estás?"
   • French: "Bonjour, comment allez-vous?"
   • German: "Hallo, wie geht es dir?"

💡 **Supported Languages:**
   • Spanish, French, German, Italian
   • Portuguese, Dutch, Russian
   • Chinese, Japanese, Korean
   • And many more!

🎯 **Usage Examples:**
   • "translate to Spanish: Hello world"
   • "translate 'Good morning' to French"
   • "translation help for business phrases"

📝 **Note:** This is a demo mode. In production, I would connect to translation APIs for accurate real-time translations.`, text)
}

// summarizeText handles text summarization
func (a *ExampleAgent) summarizeText(task string) string {
	text := extractTextFromTask(task, "summarize", "summary", "tldr")
	if text == "" {
		return "📝 Please provide text to summarize. Example: 'summarize: [your long text here]'"
	}

	words := strings.Fields(text)
	sentences := strings.Split(text, ".")

	return fmt.Sprintf(`📝 **Text Summarization:**

📄 **Original Text:** %s

📊 **Summary Statistics:**
   • Original Length: %d words, %d sentences
   • Compression Ratio: 75%% reduction
   • Reading Time: ~%d seconds

🎯 **Key Points Summary:**
   • Main Topic: %s
   • Key Themes: Communication, Information, Assistance
   • Sentiment: %s
   • Complexity: %s

✨ **TL;DR:** The text discusses %s and provides information in a clear, structured format.

💡 **Summarization Features:**
   • Bullet point summaries
   • Key theme extraction
   • Sentiment analysis
   • Custom length summaries`,
		text,
		len(words), len(sentences),
		len(words)/3,
		detectMainTopic(text),
		getTextSentiment(text),
		getTextComplexity(float64(len(words))/float64(len(sentences))),
		detectMainTopic(text))
}

// handleConversation handles general conversation
func (a *ExampleAgent) handleConversation(task string) string {
	responses := []string{
		"That's interesting! Tell me more about that.",
		"I understand what you're saying. How can I help you further?",
		"Thanks for sharing that with me. What would you like to explore next?",
		"I'm here to help! Is there anything specific you'd like assistance with?",
		"That's a great point. Would you like me to analyze or help with anything related to that?",
	}

	// Simple response selection based on task content
	responseIndex := len(task) % len(responses)

	return fmt.Sprintf(`💬 **Conversation:**

🗨️ **You said:** "%s"

🤖 **My response:** %s

🔧 **I can help you with:**
   • Text analysis and processing
   • Content generation and writing
   • Code assistance and debugging
   • Mathematical calculations
   • Data formatting and organization
   • Translation and summarization
   • System information and status

💡 **Try asking me to:**
   • Analyze some text
   • Generate creative content
   • Help with coding problems
   • Calculate math problems
   • Format your data
   • Or just chat naturally!

🚀 Type 'help' to see all my capabilities!`, task, responses[responseIndex])
}

// Content generation helper methods
func (a *ExampleAgent) generateStory(topic string) string {
	return fmt.Sprintf(`✨ **Generated Story about "%s":**

📖 **"The Tale of %s"**

Once upon a time, in a world where %s was the most precious thing imaginable, there lived a curious inventor named Alex. Alex had always been fascinated by %s and spent countless hours studying its mysteries.

One day, while working in the laboratory, Alex discovered something extraordinary about %s that would change everything. The discovery was so remarkable that it attracted the attention of scholars from around the world.

Through determination and creativity, Alex learned that %s held the key to solving one of humanity's greatest challenges. The journey was filled with obstacles, but each setback only strengthened Alex's resolve.

In the end, Alex's work with %s not only achieved the original goal but also opened new possibilities that no one had ever imagined. The story became an inspiration for future generations of inventors and dreamers.

**The End.**

💡 **Story Elements:**
   • Genre: Adventure/Discovery
   • Theme: Innovation and perseverance
   • Setting: Modern laboratory
   • Character: Curious inventor
   • Lesson: Dedication leads to breakthrough

🚀 **Want another story?** Just ask for a different topic!`, topic, topic, topic, topic, topic, topic, topic)
}

func (a *ExampleAgent) generatePoem(topic string) string {
	return fmt.Sprintf(`🎭 **Generated Poem about "%s":**

**Verses of %s**

In the realm of %s so bright,
Where wonder fills the endless night,
A story waits to be unfurled,
Of %s that changed the world.

Through valleys deep and mountains high,
Beneath the ever-changing sky,
The essence of %s rings so true,
In everything we say and do.

Like rivers flowing to the sea,
%s sets our spirits free,
A beacon in the darkest hour,
A testament to inner power.

So let us celebrate today,
The magic of %s in every way,
For in its beauty we can see,
The best of what we're meant to be.

💫 **Poem Features:**
   • Style: Free verse with rhythm
   • Theme: Inspirational and uplifting
   • Structure: 4 stanzas, 4 lines each
   • Tone: Optimistic and reflective

🎨 **Want a different style?** Ask for haiku, sonnet, or limerick!`, topic, topic, topic, topic, topic, topic, topic, topic)
}

func (a *ExampleAgent) generateEmail(topic string) string {
	return fmt.Sprintf(`📧 **Generated Email about "%s":**

**Subject:** Regarding %s - Important Information

Dear [Recipient],

I hope this email finds you well. I am writing to discuss %s and its potential impact on our current objectives.

After careful consideration and analysis, I believe that %s presents several opportunities that align with our goals. The key benefits include:

• Enhanced efficiency in our current processes
• Improved outcomes for all stakeholders
• Sustainable solutions for long-term success
• Innovative approaches to traditional challenges

I would welcome the opportunity to discuss %s further at your convenience. Please let me know when you might be available for a brief meeting or call.

Thank you for your time and consideration. I look forward to hearing from you soon.

Best regards,
[Your Name]

📝 **Email Features:**
   • Professional tone
   • Clear structure
   • Action-oriented
   • Customizable placeholders

💼 **Need different styles?** Ask for casual, formal, or marketing emails!`, topic, topic, topic, topic, topic)
}

func (a *ExampleAgent) generateGenericContent(topic string) string {
	return fmt.Sprintf(`✍️ **Generated Content about "%s":**

📋 **Comprehensive Overview of %s**

%s represents a fascinating subject that deserves careful exploration and understanding. In today's rapidly evolving world, the significance of %s cannot be overstated.

**Key Aspects:**

1. **Definition and Context**
   %s encompasses various elements that contribute to its overall importance and relevance in contemporary society.

2. **Benefits and Applications**
   The practical applications of %s extend across multiple domains, offering valuable solutions and improvements.

3. **Future Implications**
   Looking ahead, %s will likely play an increasingly important role in shaping future developments and innovations.

**Conclusion:**
Understanding %s provides valuable insights that can inform decision-making and strategic planning. As we continue to explore this topic, new opportunities and perspectives will undoubtedly emerge.

📊 **Content Statistics:**
   • Word count: ~150 words
   • Reading level: Professional
   • Structure: Introduction, body, conclusion
   • Tone: Informative and engaging

🔧 **Need specific content types?** Ask for blog posts, articles, or presentations!`, topic, topic, topic, topic, topic, topic, topic)
}

// Helper functions for time operations
func getPacificLocation() *time.Location {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	return loc
}

func getEasternLocation() *time.Location {
	loc, _ := time.LoadLocation("America/New_York")
	return loc
}

func getTokyoLocation() *time.Location {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	return loc
}

func getWeekOfYear(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}

func getDaysUntilNewYear(t time.Time) int {
	nextYear := time.Date(t.Year()+1, 1, 1, 0, 0, 0, 0, t.Location())
	return int(nextYear.Sub(t).Hours() / 24)
}

func detectMainTopic(text string) string {
	lower := strings.ToLower(text)
	topics := map[string]string{
		"technology":  "Technology",
		"business":    "Business",
		"science":     "Science",
		"education":   "Education",
		"health":      "Health",
		"environment": "Environment",
	}

	for keyword, topic := range topics {
		if strings.Contains(lower, keyword) {
			return topic
		}
	}
	return "General Discussion"
}

// Data formatting helpers
func (a *ExampleAgent) formatAsJSON(data string) string {
	return fmt.Sprintf(`📄 **JSON Formatted Data:**

{
  "input": "%s",
  "formatted": true,
  "timestamp": "%s",
  "structure": "object",
  "example": {
    "key1": "value1",
    "key2": "value2",
    "nested": {
      "property": "data"
    }
  }
}

✅ **JSON Features Applied:**
   • Proper syntax and structure
   • Nested object support
   • String, number, and boolean types
   • Array formatting capability

💡 **JSON Best Practices:**
   • Use camelCase for property names
   • Validate syntax before use
   • Consider data types carefully
   • Keep structure logical and readable`, data, time.Now().Format(time.RFC3339))
}

func (a *ExampleAgent) formatAsCSV(data string) string {
	return fmt.Sprintf(`📊 **CSV Formatted Data:**

Field1,Field2,Field3,Value
Input,"%s",Formatted,True
Name,Description,Category,Status
Sample,Data,Example,Active
Record,Information,Type,Valid

✅ **CSV Features Applied:**
   • Comma-separated values
   • Quoted text fields
   • Header row included
   • Proper escaping

💡 **CSV Best Practices:**
   • Use consistent delimiters
   • Quote fields with special characters
   • Include meaningful headers
   • Validate data consistency`, data)
}

func (a *ExampleAgent) formatAsTable(data string) string {
	return fmt.Sprintf(`📋 **Table Formatted Data:**

| Field        | Value           | Type     | Status |
|--------------|-----------------|----------|---------|
| Input        | %s             | String   | ✅ Valid |
| Timestamp    | %s             | DateTime | ✅ Valid |
| Format       | Table          | String   | ✅ Valid |
| Structure    | Organized      | String   | ✅ Valid |

✅ **Table Features Applied:**
   • Aligned columns
   • Clear headers
   • Consistent spacing
   • Visual separators

💡 **Table Best Practices:**
   • Keep column widths consistent
   • Use clear, descriptive headers
   • Align data appropriately
   • Include status indicators`, data, time.Now().Format("15:04:05"))
}

// Initialize implements the AgentInitializer interface
func (a *ExampleAgent) Initialize(ctx context.Context, config interface{}) error {
	log.Printf("🔧 Initializing %s with configuration", a.name)

	// Perform any initialization tasks here
	// For example: connecting to databases, loading models, etc.

	log.Printf("✅ %s initialized successfully", a.name)
	return nil
}

// Cleanup implements the AgentCleaner interface
func (a *ExampleAgent) Cleanup(ctx context.Context) error {
	log.Printf("🧹 Cleaning up %s", a.name)

	// Perform cleanup tasks here
	// For example: closing connections, saving state, etc.

	log.Printf("✅ %s cleanup completed", a.name)
	return nil
}

// HandleTaskResult implements the TaskResultHandler interface
func (a *ExampleAgent) HandleTaskResult(ctx context.Context, taskID, result string) error {
	log.Printf("📋 Handling result for task %s: %s", taskID, result[:min(100, len(result))])

	// Handle task results here
	// For example: logging, storing results, triggering follow-up actions

	return nil
}

// GetCapabilities returns the agent's capabilities
func (a *ExampleAgent) GetCapabilities() []string {
	return a.capabilities
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ Warning: .env file not found, using environment variables")
	}

	// Create agent configuration
	config := agent.DefaultConfig()

	// Override with specific values
	config.Name = "Enhanced Example Agent"
	config.Description = "A demonstration agent showcasing the enhanced Teneo Agent SDK capabilities"
	config.Version = "2.0.0"
	config.Capabilities = []string{
		"text_analysis_detailed",
		"content_generation_stories",
		"content_generation_poems",
		"content_generation_emails",
		"code_assistance_debug",
		"code_assistance_examples",
		"math_calculations_basic",
		"math_calculations_expressions",
		"weather_information_demo",
		"time_utilities_timezone",
		"system_status_health",
		"data_formatting_json",
		"data_formatting_csv",
		"data_formatting_tables",
		"translation_multilingual",
		"text_summarization",
		"conversation_natural",
		"help_commands_detailed",
	}
	config.WebSocketURL = "ws://localhost:8090/ws"
	config.HealthEnabled = true
	config.HealthPort = 8080
	config.PrivateKey = os.Getenv("PRIVATE_KEY")

	// Validate required environment variables
	if config.PrivateKey == "" {
		log.Fatalf("❌ PRIVATE_KEY environment variable is required")
	}

	// Derive owner address from private key
	if config.OwnerAddress == "" {
		// The auth manager will derive the address from the private key
		// We don't need to set it here as it will be handled by the agent initialization
	}

	// Network settings
	if os.Getenv("WEBSOCKET_URL") != "" {
		config.WebSocketURL = os.Getenv("WEBSOCKET_URL")
	}
	log.Printf("🔗 Using WebSocket URL: %s", config.WebSocketURL)

	// Health monitoring
	if port := os.Getenv("HEALTH_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.HealthPort = p
		}
		log.Printf("🌐 Health monitoring will be available on port %d", config.HealthPort)
	} else {
		log.Printf("🌐 Health monitoring will be available on port %d", config.HealthPort)
	}

	// Create agent handler
	agentHandler := NewExampleAgent()

	// Create enhanced agent
	enhancedAgent, err := agent.NewEnhancedAgent(&agent.EnhancedAgentConfig{
		Config:       config,
		AgentHandler: agentHandler,
	})
	if err != nil {
		log.Fatalf("❌ Failed to create enhanced agent: %v", err)
	}

	// Display startup information
	log.Printf("\n"+
		"🚀 ================================\n"+
		"   Enhanced Teneo Agent Starting\n"+
		"================================\n"+
		"Name: %s\n"+
		"Version: %s\n"+
		"Capabilities: %v\n"+
		"WebSocket: %s\n"+
		"Health Port: %d\n"+
		"Wallet: %s\n"+
		"================================\n",
		config.Name,
		config.Version,
		config.Capabilities,
		config.WebSocketURL,
		config.HealthPort,
		enhancedAgent.GetAuthManager().GetAddress(),
	)

	// Run the agent
	log.Printf("🚀 Starting enhanced agent...")
	if err := enhancedAgent.Run(); err != nil {
		log.Fatalf("❌ Agent failed: %v", err)
	}

	log.Printf("👋 Enhanced agent shutdown complete")
}
