package engines

import (
	"context"
	"fmt"
	"io"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAI TTS example usage
// https://github.com/sashabaranov/go-openai/pull/617/files

type OpenAIEngine struct {
	client *openai.Client
	model  openai.SpeechModel
	voice  openai.SpeechVoice
}

func NewOpenAIEngine(apiKey string, model openai.SpeechModel, voice openai.SpeechVoice) *OpenAIEngine {
	client := openai.NewClient(apiKey) // Ensure the openai package is correctly imported
	return &OpenAIEngine{
		client: client,
		model:  model,
		voice:  voice,
	}
}

func (o *OpenAIEngine) Synthesize(text string) bool {
	ctx := context.Background()

	req := openai.CreateSpeechRequest{
		Model:          o.model,
		Input:          text,
		Voice:          o.voice,
		ResponseFormat: openai.SpeechResponseFormatMp3,
		Speed:          1.0,
	}

	resp, err := o.client.CreateSpeech(ctx, req)
	if err != nil {
		fmt.Printf("Speech synthesis error: %v\n", err)
		return false
	}

	// Assuming you want to write the response to a file
	out, err := os.Create("output.mp3")
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return false
	}
	defer out.Close()

	_, err = io.Copy(out, resp)
	if err != nil {
		fmt.Printf("Error writing to output file: %v\n", err)
		return false
	}

	return true
}
