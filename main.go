package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Config holds the configuration for the Tautulli API.
type Config struct {
	TautulliURL    string
	TautulliAPIKey string
}

// TautulliResponse defines the structure for the JSON response from the Tautulli API.
type TautulliResponse struct {
	Response struct {
		Data struct {
			StreamCount string    `json:"stream_count"`
			Sessions    []Session `json:"sessions"`
		} `json:"data"`
	} `json:"response"`
}

// Session represents a single media stream from the Tautulli API.
type Session struct {
	User             string `json:"user"`
	Player           string `json:"player"`
	GrandparentTitle string `json:"grandparent_title"`
	Title            string `json:"title"`
	MediaType        string `json:"media_type"`
	Thumb            string `json:"thumb"`
	ProgressPercent  string `json:"progress_percent"`
	PosterURL        string // This will be constructed in our code
	Progress         int    // This will be calculated
}

// PageData is the data structure passed to the HTML template.
type PageData struct {
	StreamCount int
	Sessions    []Session
	Timestamp   string
}

// htmlTemplate has been completely redesigned for a horizontal, card-based grid layout.
const htmlTemplate = `
<div class="view view--full" style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;">
    <div class="layout">
        <div class="column">
            <div class="markdown gap--medium">
                {{if gt .StreamCount 0}}
                    <div style="display: flex; flex-wrap: wrap; gap: 1rem; justify-content: flex-start;">
                        {{range .Sessions}}
                            <div style="flex: 1 1 48%; min-width: 350px; display: flex; background-color: #f9f9f9; border-radius: 8px; overflow: hidden; border: 1px solid #eee;">
                                <div style="flex-shrink: 0;">
                                    <img src="{{.PosterURL}}" style="width: 120px; height: 180px; object-fit: cover;" alt="Poster" />
                                </div>
                                <div style="padding: 0.75rem 1rem; display: flex; flex-direction: column; justify-content: center; flex-grow: 1;">
                                    <p style="margin: 0; font-weight: bold; font-size: 1.2rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">
                                        {{if eq .MediaType "episode"}}
                                            {{.GrandparentTitle}}
                                        {{else}}
                                            {{.Title}}
                                        {{end}}
                                    </p>
                                    {{if eq .MediaType "episode"}}
                                        <p style="margin: 0.25rem 0; font-size: 1rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">{{.Title}}</p>
                                    {{end}}
                                    <p style="margin: 0.5rem 0 0; font-size: 0.9rem; color: #555;">
                                        {{.User}}
                                    </p>
									<div style="margin-top: auto; padding-top: 1rem;">
										<div style="background-color: #e0e0e0; border-radius: 4px; overflow: hidden;">
											<div style="height: 8px; width: {{.Progress}}%; background-color: #76c7c0; border-radius: 4px;"></div>
										</div>
									</div>
                                </div>
                            </div>
                        {{end}}
                    </div>
                {{else}}
                    <div class="content-element content content--center">
                        <p>Nothing is currently playing.</p>
                    </div>
                {{end}}
                <span class="label label--underline mt-4" style="text-align: right; display: block;">Updated: {{.Timestamp}}</span>
            </div>
        </div>
    </div>
</div>
`

// httpHandler fetches data and renders the richer HTML layout.
func httpHandler(w http.ResponseWriter, r *http.Request, config *Config) {
	// 1. Construct the Tautulli API URL.
	apiURL := fmt.Sprintf("%s/api/v2?apikey=%s&cmd=get_activity", config.TautulliURL, config.TautulliAPIKey)

	// 2. Make the request to Tautulli.
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		http.Error(w, "Failed to connect to Tautulli", http.StatusInternalServerError)
		log.Printf("Error connecting to Tautulli: %v", err)
		return
	}
	defer resp.Body.Close()

	// 3. Decode the JSON response.
	var tautulliData TautulliResponse
	if err := json.NewDecoder(resp.Body).Decode(&tautulliData); err != nil {
		http.Error(w, "Failed to parse Tautulli response", http.StatusInternalServerError)
		log.Printf("Error parsing JSON from Tautulli: %v", err)
		return
	}

	// 4. Convert stream_count to an integer.
	streamCount, err := strconv.Atoi(tautulliData.Response.Data.StreamCount)
	if err != nil {
		http.Error(w, "Failed to parse stream count", http.StatusInternalServerError)
		log.Printf("Error converting stream_count: %v", err)
		return
	}

	// Limit to a maximum of 4 sessions for the display
	sessions := tautulliData.Response.Data.Sessions
	if len(sessions) > 4 {
		sessions = sessions[:4]
	}

	// 5. Construct poster URLs and calculate progress for each session.
	for i := range sessions {
		session := &sessions[i]
		if session.Thumb != "" {
			encodedThumb := url.QueryEscape(session.Thumb)
			session.PosterURL = fmt.Sprintf("%s/api/v2?apikey=%s&cmd=pms_image_proxy&img=%s", config.TautulliURL, config.TautulliAPIKey, encodedThumb)
		} else {
			session.PosterURL = "https://placehold.co/120x180/eee/ccc?text=No+Art"
		}

		// Calculate progress
		if progress, err := strconv.Atoi(session.ProgressPercent); err == nil {
			session.Progress = progress
		}
	}

	// 6. Prepare data for the template.
	pageData := PageData{
		StreamCount: streamCount,
		Sessions:    sessions,
		Timestamp:   time.Now().Format("3:04 PM"),
	}

	// 7. Parse and execute the template.
	tmpl, err := template.New("trmnl").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, "Failed to parse HTML template", http.StatusInternalServerError)
		log.Printf("Error parsing template: %v", err)
		return
	}

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

	if config.TautulliURL == "" || config.TautulliAPIKey == "" {
		log.Fatal("TAUTULLI_URL and TAUTULLI_API_KEY environment variables must be set.")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpHandler(w, r, config)
	})

	port := "8080"
	log.Printf("Starting Tautulli TRMNL plugin server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
