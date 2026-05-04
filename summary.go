package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// OllamaRequest matches the expected JSON payload for Ollama's API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse matches the JSON response from Ollama
type OllamaResponse struct {
	Response string `json:"response"`
}

const (
	ollamaURL = "http://localhost:11434/api/generate"
	modelName = "qwen3:8b" // Change this if you want to use a different model (e.g., mistral, phi3)
)

// summarizeEmail sends the text to Ollama and returns the summary
func summarizeEmail(content string) (string, error) {
	prompt := fmt.Sprintf("Please provide a concise summary of the following email:\n\n%s", content)

	reqBody := OllamaRequest{
		Model:  modelName,
		Prompt: prompt,
		Stream: false, // We want the full response at once, not streamed
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// Send POST request to local Ollama API
	resp, err := http.Post(ollamaURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to connect to Ollama (is it running?): %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ollamaResp OllamaResponse
	err = json.Unmarshal(bodyBytes, &ollamaResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse Ollama response: %v", err)
	}

	return strings.TrimSpace(ollamaResp.Response), nil
}

func main() {
	sourceDir := "./my_emails" // Folder where your email .txt files are
	destDir := "./summaries"   // Folder where summaries will be saved

	// 1. Create the summaries directory if it doesn't exist
	err := os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create destination directory: %v", err)
	}

	// 2. Read all files in the source directory
	files, err := os.ReadDir(sourceDir)
	if err != nil {
		log.Fatalf("Failed to read directory %s: %v", sourceDir, err)
	}

	// 3. Process each file one at a time
	for _, file := range files {
		// Skip directories and only process .txt files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".txt") {
			continue
		}

		filePath := filepath.Join(sourceDir, file.Name())
		fmt.Printf("Summarizing: %s...\n", file.Name())

		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading file %s: %v\n", filePath, err)
			continue
		}

		// Call Ollama
		summary, err := summarizeEmail(string(content))
		if err != nil {
			log.Printf("Error summarizing %s: %v\n", filePath, err)
			continue
		}

		// 4. Save the summary
		destPath := filepath.Join(destDir, "summary_"+file.Name())
		err = os.WriteFile(destPath, []byte(summary), 0644)
		if err != nil {
			log.Printf("Error writing summary for %s: %v\n", filePath, err)
		} else {
			fmt.Printf("Saved: %s\n\n", destPath)
		}
	}

	fmt.Println("All emails processed.")
}
