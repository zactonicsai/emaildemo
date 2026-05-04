package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Opens a URL specifically in Google Chrome
func openInChrome(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Use PowerShell to safely pass the URL containing '&' characters
		cmd = exec.Command("powershell", "-Command", fmt.Sprintf("Start-Process chrome -ArgumentList '%s'", url))
	case "darwin": // macOS
		cmd = exec.Command("open", "-a", "Google Chrome", url)
	case "linux":
		cmd = exec.Command("google-chrome", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

// Retrieves a token by opening Chrome and starting a local server for the callback
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Println("Opening Google Chrome to authenticate...")
	err := openInChrome(authURL)
	if err != nil {
		fmt.Printf("Could not open Chrome automatically. Please open this link manually:\n%v\n", authURL)
	}

	// Channel to wait for the authorization code
	codeCh := make(chan string)
	srv := &http.Server{Addr: ":8080"}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			fmt.Fprintf(w, "Authentication successful! You can close this window and return to your terminal.")
			codeCh <- code
		} else {
			fmt.Fprintf(w, "Failed to get authorization code.")
		}
	})

	// Start local server to catch the redirect
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// Expected error when shutting down the server
		}
	}()

	// Block until we receive the code
	authCode := <-codeCh

	// Shutdown the server cleanly
	srv.Shutdown(context.Background())

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token, saves it, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// Recursively parse the email payload to find the plain text body
func extractBody(payload *gmail.MessagePart) string {
	if payload.MimeType == "text/plain" && payload.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err != nil {
			// Gmail sometimes uses raw URL encoding (no padding)
			data, _ = base64.RawURLEncoding.DecodeString(payload.Body.Data)
		}
		return string(data)
	}

	for _, part := range payload.Parts {
		if body := extractBody(part); body != "" {
			return body
		}
	}
	return ""
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <destination_folder>\n", filepath.Base(os.Args[0]))
	}

	destDir := os.Args[1]
	err := os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create destination directory: %v", err)
	}

	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file (ensure credentials.json is in this folder): %v", err)
	}

	// Must match the redirect URI set in Google Cloud Console
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	config.RedirectURL = "http://localhost:8080/callback"

	client := getClient(config)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	fmt.Println("Fetching recent emails...")

	// Limiting to 10 messages for testing. Change MaxResults to fetch more.
	r, err := srv.Users.Messages.List("me").MaxResults(10).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve messages: %v", err)
	}

	if len(r.Messages) == 0 {
		fmt.Println("No messages found.")
		return
	}

	for _, m := range r.Messages {
		msg, err := srv.Users.Messages.Get("me", m.Id).Format("full").Do()
		if err != nil {
			log.Printf("Error retrieving message %s: %v", m.Id, err)
			continue
		}

		body := extractBody(msg.Payload)
		if body == "" {
			body = "[No plain text body found in this email or email is purely HTML]"
		}

		// Extract Subject for metadata
		subject := "No Subject"
		for _, header := range msg.Payload.Headers {
			if header.Name == "Subject" {
				subject = header.Value
				break
			}
		}

		content := fmt.Sprintf("Subject: %s\nMessage ID: %s\nSnippet: %s\n\nBody:\n%s", subject, msg.Id, msg.Snippet, body)

		filePath := filepath.Join(destDir, fmt.Sprintf("%s.txt", msg.Id))
		err = os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			log.Printf("Failed to write file %s: %v", filePath, err)
		} else {
			fmt.Printf("Saved: %s\n", filePath)
		}
	}
	fmt.Println("Done!")
}
