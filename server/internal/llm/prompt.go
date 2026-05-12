package llm

import "fmt"

// BuildPrompt builds a compact LLM prompt from session context.
func BuildPrompt(language, summary, latestText string, notes []string) string {
	prompt := `You are a realtime speech AI assistant.

Language:
%s

Conversation summary:
%s

Important notes:
%s

Latest user message:
%s

Instruction:
Answer shortly, clearly, and in the same language as the user.
Avoid long explanations unless the user asks for details.`

	notesStr := "None"
	if len(notes) > 0 {
		notesStr = ""
		for _, n := range notes {
			notesStr += "- " + n + "\n"
		}
	}

	if summary == "" {
		summary = "No previous context."
	}

	return fmt.Sprintf(prompt, language, summary, notesStr, latestText)
}

// BuildLanguageDetectPrompt builds a prompt for AI-based language detection.
func BuildLanguageDetectPrompt(text string) string {
	return fmt.Sprintf(`Detect the language of the following text. 
Reply with ONLY the ISO 639-1 language code (e.g. "en", "vi", "ja", "ko", "zh", "fr", "de", "es") and nothing else.

Text: %s`, text)
}
