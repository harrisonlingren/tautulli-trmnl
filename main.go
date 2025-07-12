package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

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

const htmlTemplate = `<markup>
<div class="view view--full">
    <div class="layout">
        <div class="column">
            {{if gt .StreamCount 0}}
            <div class="layout mt-4">
                <div class="columns columns--2">
                    {{range .Sessions}}
                    <div class="column">
                        <div class="widget">
                            <div class="widget__media">
                                <img src="{{.PosterURL}}" alt="Poster" />
                            </div>
                            <div class="widget__body">
                                <span class="widget__title">
                                    {{if eq .MediaType "episode"}}
                                        {{.GrandparentTitle}}
                                    {{else}}
                                        {{.Title}}
                                    {{end}}
                                </span>
                                {{if eq .MediaType "episode"}}
                                    <span class="widget__subtitle">{{.Title}}</span>
                                {{end}}
                                <span class="widget__label mt-2">{{.User}}</span>
                                <div class="progress-bar mt-3">
                                    <div class="progress-bar__fg" style="width: {{.Progress}}%;"></div>
                                </div>
                            </div>
                        </div>
                    </div>
                    {{end}}
                </div>
            </div>
            {{else}}
            <div class="markdown">
                <div class="content-element content content--center mt-4">
                    <p>Nothing is currently playing.</p>
                </div>
            </div>
            {{end}}
            <span class="label label--underline mt-4">Updated: {{.Timestamp}}</span>
        </div>
    </div>
</div>
</markup>
`

// imageProxyHandler fetches images from Tautulli and serves them through our local server.
func imageProxyHandler(w http.ResponseWriter, r *http.Request) {
	tautulliURL := r.URL.Query().Get("tautulli_url")
	apiKey := r.URL.Query().Get("api_key")
	imgPath := r.URL.Query().Get("img")

	if tautulliURL == "" || apiKey == "" || imgPath == "" {
		http.Error(w, "Missing required query parameters for image proxy", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(tautulliURL, "http://") && !strings.HasPrefix(tautulliURL, "https://") {
		tautulliURL = "https://" + tautulliURL
	}

	// Construct the full, original Tautulli image proxy URL.
	fullImgURL := fmt.Sprintf("%s/api/v2?apikey=%s&cmd=pms_image_proxy&img=%s", tautulliURL, apiKey, imgPath)

	// Fetch the image from Tautulli.
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fullImgURL)
	if err != nil {
		http.Error(w, "Failed to fetch image from Tautulli", http.StatusInternalServerError)
		log.Printf("Image proxy failed to connect to Tautulli: %v", err)
		return
	}
	defer resp.Body.Close()

	// Copy the headers from the Tautulli response (like Content-Type) to our response.
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Stream the image data directly to the client.
	io.Copy(w, resp.Body)
}

// httpHandler fetches data and renders the richer HTML layout.
func httpHandler(w http.ResponseWriter, r *http.Request) {
	// Get Tautulli URL and API Key from query parameters.
	tautulliURL := r.URL.Query().Get("tautulli_url")
	apiKey := r.URL.Query().Get("api_key")

	if tautulliURL == "" || apiKey == "" {
		http.Error(w, "Missing required query parameters: 'tautulli_url' and 'api_key'", http.StatusBadRequest)
		log.Println("Error: Received request with missing 'tautulli_url' or 'api_key' query parameters.")
		return
	}

	if !strings.HasPrefix(tautulliURL, "http://") && !strings.HasPrefix(tautulliURL, "https://") {
		tautulliURL = "https://" + tautulliURL
	}

	// 1. Construct the Tautulli API URL.
	apiURL := fmt.Sprintf("%s/api/v2?apikey=%s&cmd=get_activity", tautulliURL, apiKey)

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
		// If stream_count is empty or not a number, default to 0.
		streamCount = 0
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
			session.PosterURL = fmt.Sprintf("/image?img=%s&tautulli_url=%s&api_key=%s", encodedThumb, url.QueryEscape(tautulliURL), url.QueryEscape(apiKey))
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
	http.HandleFunc("/", httpHandler)
	http.HandleFunc("/image", imageProxyHandler)

	port := "8080"
	log.Printf("Starting Tautulli TRMNL plugin server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
