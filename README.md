# Tautulli Activity Plugin for TRMNL

This plugin creates a web service that connects to your [Tautulli](https://tautulli.com/) instance to display the current Plex Media Server activity on a [TRMNL](https://usetrmnl.com/) display. It provides a rich, graphical overview of currently playing media, inspired by the Tautulli web interface.

## Preview

The plugin renders a flexible, card-based layout that can display up to four concurrent streams.

![Plugin Preview](https://i.imgur.com/gY8g2pC.png)

## Features

-   **Graphical Layout:** Displays media posters for a clean, visual overview.
-   **Real-time Activity:** Shows who is watching what, and on which device.
-   **Playback Progress:** Includes progress bars to see how far along each stream is.
-   **Flexible Grid:** Displays 1-4 streams in a responsive grid that adapts to the number of active sessions.
-   **Lightweight:** Built in Go as a single, self-contained binary with no external dependencies needed at runtime.

## Setup Instructions

Follow these steps to get the plugin running and connected to your TRMNL account.

### Prerequisites

1.  A running **Tautulli** instance connected to your Plex server.
2.  **Go** (version 1.18 or later) installed on the machine where you will run this service.
3.  The service must be accessible from the internet (e.g., running on a public server or exposed via a tool like `ngrok` for testing).

### 1. Get Your Tautulli API Key

-   Open your Tautulli web interface.
-   Go to **Settings** > **Web Interface**.
-   Scroll down to the **API** section.
-   Make sure **Enable API** is checked.
-   Copy the **API Key**.

### 2. Run the Plugin Service

1.  **Download the Code:**
    Save the Go code into a file named `main.go`.

2.  **Start the Service:**
    Open your terminal, navigate to the directory where you saved `main.go`, and run:
    ```bash
    go run main.go
    ```
    You should see the message: `Starting Tautulli TRMNL plugin server on port 8080`.

### 3. Configure TRMNL

1.  **Create a Private Plugin:**
    -   Log in to your TRMNL account.
    -   Navigate to **Plugins** > **Private Plugin** and create a new one.

2.  **Set the Polling Strategy:**
    -   **Strategy:** Choose **Polling**.
    -   **Polling URL:** Enter the public URL of the Go service, providing your Tautulli URL and API key as query parameters.

    Your URL must be in the following format:
    ```
    http://YOUR_SERVER_IP:8080/?tautulli_url=YOUR_TAUTULLI_URL&api_key=YOUR_API_KEY
    ```

    **Example:**
    ```
    [http://123.45.67.89:8080/?tautulli_url=http://192.168.1.100:8181&api_key=abcdef1234567890](http://123.45.67.89:8080/?tautulli_url=http://192.168.1.100:8181&api_key=abcdef1234567890)
    ```
    > **Note:** Make sure to replace the example values with your actual server IP, Tautulli URL, and Tautulli API Key.

3.  **Add to a Playlist:**
    -   Save the plugin.
    -   Add your new private plugin to one of your TRMNL device's playlists.
    -   Force a refresh on the device, and you should see your Plex activity appear!
