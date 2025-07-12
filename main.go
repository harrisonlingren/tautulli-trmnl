package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

// Config holds the configuration for the Tautulli API.
// These values are loaded from environment variables for security and flexibility.
type Config struct {
	TautulliURL    string
	TautulliAPIKey string
}

// TautulliResponse defines the structure for the JSON response from the Tautulli API.
// We only map the fields we need for the plugin.
type TautulliResponse struct {
	Response struct {
		Data struct {
			StreamCount int       `json:"stream_count"`
			Sessions    []Session `json:"sessions"`
		} `json:"data"`
	} `json:"response"`
}

// Session represents a single media stream currently being played on Plex.
type Session struct {
	User             string `json:"user"`
	Player           string `json:"player"`
	GrandparentTitle string `json:"grandparent_title"` // e.g., The name of the TV Show
	Title            string `json:"title"`            // e.g., The movie title or episode title
	MediaType        string `json:"media_type"`       // "movie" or "episode"
	Poster           string `json:"poster"`           // URL for the media poster
}

// PageData is the data structure passed to the HTML template for rendering.
type PageData struct {
	StreamCount int
	Sessions    []Session
	Timestamp   string
}

// htmlTemplate is the TRMNL-compatible HTML that will be rendered.
// It uses Liquid-style template syntax to display the data.
const htmlTemplate = `
<div class="view view--full">
    <div class="layout">
        <div class="column">
            <div class="markdown gap--large">
                <span class="title">Plex Activity</span>
                {{if gt .StreamCount 0}}
                    {{range .Sessions}}
                        <div class="content-element content" style="border-bottom: 1px solid #eee; padding-bottom: 1rem; margin-bottom: 1rem;">
                            <p><strong>{{.User}}</strong> on {{.Player}}</p>
                            {{if eq .MediaType "episode"}}
                                <p><em>{{.GrandparentTitle}}</em> - {{.Title}}</p>
                            {{else}}
                                <p><strong>{{.Title}}</strong></p>
                            {{end}}
                        </div>
                    {{end}}
                {{else}}
                    <div class="content-element content content--center">
                        <p>Nothing is currently playing.</p>
                    </div>
                {{end}}
                <span class="label label--underline mt-4">Updated: {{.Timestamp}}</span>
            </div>
        </div>
    </div>
</div>
`

// httpHandler is the main function that handles incoming requests from TRMNL.
func httpHandler(w http.ResponseWriter, r *http.Request, config *Config) {
	// 1. Construct the Tautulli API URL from the configuration.
	apiURL := fmt.Sprintf("%s/api/v2?apikey=%s&cmd=get_activity", config.TautulliURL, config.TautulliAPIKey)

	// 2. Create a new HTTP client with a timeout to prevent hanging requests.
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		http.Error(w, "Failed to connect to Tautulli", http.StatusInternalServerError)
		log.Printf("Error connecting to Tautulli at %s: %v", config.TautulliURL, err)
		return
	}
	defer resp.Body.Close()

	// 3. Decode the JSON response into our Go structs.
	var tautulliData TautulliResponse
	if err := json.NewDecoder(resp.Body).Decode(&tautulliData); err != nil {
		http.Error(w, "Failed to parse Tautulli response", http.StatusInternalServerError)
		log.Printf("Error parsing JSON from Tautulli: %v", err)
		return
	}

	// 4. Prepare the data structure to be passed to the HTML template.
	pageData := PageData{
		StreamCount: tautulliData.Response.Data.StreamCount,
		Sessions:    tautulliData.Response.Data.Sessions,
		Timestamp:   time.Now().Format("3:04 PM"),
	}

	// 5. Parse our HTML template string.
	tmpl, err := template.New("trmnl").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, "Failed to parse HTML template", http.StatusInternalServerError)
		log.Printf("Error parsing internal HTML template: %v", err)
		return
	}

	// 6. Set the content type and execute the template, writing the output to the response.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, pageData); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Error executing template: %v", err)
	}
}

func main() {
	// Load configuration from environment variables.
	config := &Config{
		TautulliURL:    os.Getenv("TAUTULLI_URL"),
		TautulliAPIKey: os.Getenv("TAUTULLI_API_KEY"),
	}

	// Ensure that the required configuration is present.
	if config.TautulliURL == "" || config.TautulliAPIKey == "" {
		log.Fatal("TAUTULLI_URL and TAUTULLI_API_KEY environment variables must be set.")
	}

	// Register our handler function for the root path.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpHandler(w, r, config)
	})

	// Start the web server.
	port := "8080"
	log.Printf("Starting Tautulli TRMNL plugin server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
