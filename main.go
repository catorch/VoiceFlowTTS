package main

import (
	"fmt"
	"log"
	"os"
	"stream/engines"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get the API key from the environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("API key not set in .env file")
	}

	// Check if a text argument is provided
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go \"<text to synthesize>\"")
	}
	inputText := os.Args[1]

	engine := engines.NewOpenAIEngine(
		apiKey,
		openai.TTSModel1,
		openai.VoiceAlloy,
	)

	success := engine.SynthesizeV3(inputText)
	if success {
		fmt.Println("Synthesis successful, output saved to output.mp3")
	} else {
		fmt.Println("Synthesis failed")
	}
}
