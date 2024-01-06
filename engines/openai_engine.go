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

func (o *OpenAIEngine) SynthesizeV2(text string) bool {
	ctx := context.Background()

	// Creating a Chat Completion request with the received text
	chatResp, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: text,
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return false
	}

	// Extracting the response content
	generatedText := chatResp.Choices[0].Message.Content

	req := openai.CreateSpeechRequest{
		Model:          o.model,
		Input:          generatedText,
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

	decoder, err := mp3.NewDecoder(resp)
	if err != nil {
		fmt.Printf("Error decoding MP3: %v\n", err)
		return false
	}

	context, err := oto.NewContext(decoder.SampleRate(), 2, 2, 8192)
	if err != nil {
		fmt.Printf("Error creating audio context: %v\n", err)
		return false
	}
	defer context.Close()

	player := context.NewPlayer()
	defer player.Close()

	audioChunks := make(chan []byte, 10)
	done := make(chan bool)

	// Goroutine for decoding and sending audio chunks
	go func() {
		defer close(audioChunks)
		for {
			buf := make([]byte, 1024)
			n, err := decoder.Read(buf)
			if err != nil && err != io.EOF {
				fmt.Printf("Error while decoding: %v\n", err)
				done <- false
				return
			}
			if n == 0 {
				break
			}
			audioChunks <- buf[:n]
		}
		done <- true
	}()

	// Goroutine for playing audio
	go func() {
		for chunk := range audioChunks {
			if _, err := player.Write(chunk); err != nil {
				fmt.Printf("Error playing audio: %v\n", err)
				done <- false
				return
			}
		}
		done <- true
	}()

	for i := 0; i < 2; i++ {
		if !<-done {
			return false
		}
	}

	return true
}
