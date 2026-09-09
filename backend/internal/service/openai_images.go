package service

import (
	"fmt"
	"strings"
)

// OpenAIImagesRequest represents a request to the OpenAI-compatible images endpoint.
type OpenAIImagesRequest struct {
	Model          string `json:"model" binding:"required"`
	Prompt         string `json:"prompt" binding:"required,max=4000"`
	N              int    `json:"n" binding:"min=1,max=10"`
	Size           string `json:"size" binding:"oneof=256x256 512x512 1024x1024 1792x1024 1024x1792"`
	Quality        string `json:"quality" binding:"oneof=standard hd"`
	ResponseFormat string `json:"response_format" binding:"oneof=url b64_json"`
	Style          string `json:"style" binding:"oneof=natural vivid"`
	Stream         bool   `json:"stream"`
}

// Validate performs additional validation beyond struct tags.
func (r *OpenAIImagesRequest) Validate() error {
	if strings.TrimSpace(r.Prompt) == "" {
		return fmt.Errorf("prompt cannot be empty")
	}
	return nil
}

// OpenAIImagesResponse is the OpenAI-compatible response format.
type OpenAIImagesResponse struct {
	Created int64              `json:"created"`
	Data    []OpenAIImagesData `json:"data"`
}

// OpenAIImagesData represents a single image result.
type OpenAIImagesData struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// OpenAIImagesError represents an error response.
type OpenAIImagesError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// ModelInfo describes an available image model.
type ModelInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Provider     string   `json:"provider"`
	Capabilities []string `json:"capabilities"`
	Sizes        []string `json:"supported_sizes"`
	Formats      []string `json:"supported_formats"`
}

// AvailableModels returns the list of supported image models.
func AvailableModels() []ModelInfo {
	return []ModelInfo{
		{
			ID:           "gpt-image-1",
			Name:         "GPT Image 1",
			Provider:     "openai",
			Capabilities: []string{"generation", "edit"},
			Sizes:        []string{"1024x1024", "1792x1024", "1024x1792"},
			Formats:      []string{"png", "jpeg", "webp"},
		},
		{
			ID:           "dall-e-3",
			Name:         "DALL-E 3",
			Provider:     "openai",
			Capabilities: []string{"generation"},
			Sizes:        []string{"1024x1024", "1792x1024", "1024x1792"},
			Formats:      []string{"png"},
		},
	}
}
