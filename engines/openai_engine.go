package engines

import (
	"context"
	"fmt"
	"io"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
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
	defer resp.Close()

	// Decode MP3 data
	decoder, err := mp3.NewDecoder(resp)
	if err != nil {
		fmt.Printf("Error decoding MP3: %v\n", err)
		return false
	}

	// Initialize audio context and player
	context, err := oto.NewContext(decoder.SampleRate(), 2, 2, 8192)
	if err != nil {
		fmt.Printf("Error creating audio context: %v\n", err)
		return false
	}
	defer context.Close()

	player := context.NewPlayer()
	defer player.Close()

	// Play audio
	if _, err := io.Copy(player, decoder); err != nil {
		fmt.Printf("Error playing audio: %v\n", err)
		return false
	}

	return true
}
