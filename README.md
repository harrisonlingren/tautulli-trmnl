# Tautulli Activity Plugin for TRMNL

This project provides a simple, text-based TRMNL plugin to display current Plex Media Server activity by fetching data from your [Tautulli](https://tautulli.com/) instance.

It uses a decoupled architecture:
1.  A lightweight **Go web service** that acts as a JSON API, fetching data from Tautulli.
2.  A **Liquid markup template** that you paste into the TRMNL private plugin editor to render the JSON data.

This approach separates the data logic from the presentation, making the plugin robust and easy to customize.

## Preview

The final plugin displays up to four concurrent streams in a clean, row-based layout optimized for the TRMNL e-ink display. Each row shows the media title, user, and a progress bar. A title bar at the bottom displays the plugin name and the last update time.

*(Note: The preview below is a representation of the final layout.)*

![Plugin Preview](https://i.imgur.com/kG8E1qW.png)

## Features

-   **Decoupled Architecture:** Separates the Go backend (data) from the Liquid frontend (presentation).
-   **Text-Optimized Layout:** A clean, row-based layout that is highly readable on e-ink displays.
-   **At-a-Glance Info:** Displays media title, series/episode title, user, and playback progress.
-   **TRMNL v2 Compliant:** Uses official framework components for the grid layout and title bar.
-   **Lightweight:** The Go service is a single, self-contained binary with no external runtime dependencies.

## Setup Instructions

This plugin requires a two-part setup: running the Go service and configuring your TRMNL private plugin.

### Part 1: Run the Go Service

This service fetches data from Tautulli and provides it as a JSON endpoint.

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/harrisonlingren/tautulli-trmnl.git
    cd tautulli-trmnl
    ```

2.  **Run the Service:**
    From the project directory, run the application. It will start a web server on port `8080`.
    ```bash
    go run main.go
    ```
    You should see the message: `Starting Tautulli TRMNL plugin server on port 8080`.

3.  **Expose the Service:**
    The Go service must be accessible from the internet. For local testing, a tool like [ngrok](https://ngrok.com/) is recommended. For permanent use, you should deploy it to a public server.
    ```bash
    # Example using ngrok
    ngrok http 8080
    ```
    Ngrok will give you a public URL (e.g., `https://random-string.ngrok.io`). This is your `YOUR_SERVER_URL`.

### Part 2: Configure TRMNL Private Plugin

1.  **Create a Private Plugin:**
    -   Log in to your TRMNL account.
    -   Navigate to **Plugins** > **Private Plugin** and create a new one.

2.  **Set the Polling URL:**
    -   **Strategy:** Choose **Polling**.
    -   **Polling URL:** Construct a URL using your public server URL from the previous step, and your Tautulli details.

    The format is:
    `YOUR_SERVER_URL/?tautulli_url=YOUR_TAUTULLI_URL&api_key=YOUR_API_KEY`

    **Example:**
    `https://random-string.ngrok.io/?tautulli_url=http://192.168.1.100:8181&api_key=abcdef1234567890`

3.  **Add the Markup:**
    -   In the TRMNL plugin editor, paste the entire block of code from `full.liquid`, `half_horizontal.liquid`, `half-vertical.liquid`, or `quadrant.liquid` to meet your desired layout types.

4.  **Save and Add to Playlist:**
    -   Save the private plugin.
    -   Add it to your TRMNL device's playlist.
    -   Force a refresh on the device to see the result.
