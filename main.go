package main

import (
	"encoding/json"
	"fmt"
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
	PosterURL        string `json:"poster_url"` // This will be constructed in our code
	Progress         int    `json:"progress"`   // This will be calculated
}

// PageData is the root object for our JSON response.
type PageData struct {
	StreamCount int       `json:"stream_count"`
	Sessions    []Session `json:"sessions"`
	Timestamp   string    `json:"timestamp"`
}

// httpHandler fetches data from Tautulli and returns it as a JSON object.
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
		streamCount = 0
	}

	// Limit to a maximum of 4 sessions for the display
	sessions := tautulliData.Response.Data.Sessions
	if len(sessions) > 4 {
		sessions = sessions[:4]
	}

	// 5. Construct full poster URLs and calculate progress for each session.
	for i := range sessions {
		session := &sessions[i]
		if session.Thumb != "" {
			// **MODIFIED:** Create the full, absolute URL for the poster.
			encodedThumb := url.QueryEscape(session.Thumb)
			session.PosterURL = fmt.Sprintf("%s/api/v2?apikey=%s&cmd=pms_image_proxy&img=%s", tautulliURL, apiKey, encodedThumb)
		} else {
			session.PosterURL = "https://placehold.co/120x180/eee/ccc?text=No+Art"
		}

		if progress, err := strconv.Atoi(session.ProgressPercent); err == nil {
			session.Progress = progress
		}
	}

	// 6. Prepare data for the final JSON response.
	pageData := PageData{
		StreamCount: streamCount,
		Sessions:    sessions,
		Timestamp:   time.Now().Format("3:04 PM"),
	}

	// 7. Set the content type and encode the response as JSON.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pageData); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func main() {
	http.HandleFunc("/", httpHandler)

	port := "8080"
	log.Printf("Starting Tautulli TRMNL plugin server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
