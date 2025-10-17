package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type Video struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Published time.Time `json:"published"`
	URL       string    `json:"url"`
	Year      int       `json:"year"`
}

type YouTubeSearchResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			Title       string `json:"title"`
			PublishedAt string `json:"publishedAt"`
		} `json:"snippet"`
	} `json:"items"`
	NextPageToken string `json:"nextPageToken"`
}

func main() {
	// Load .env file manually
	loadEnvFile("../.env")

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY not set in .env or environment")
	}

	channelID := os.Getenv("CHANNEL_ID")
	if channelID == "" {
		log.Fatal("CHANNEL_ID not set in .env or environment")
	}

	videos := []Video{}

	// Fetch all videos from the channel
	pageToken := ""
	for {
		url := fmt.Sprintf(
			"https://www.googleapis.com/youtube/v3/search?key=%s&channelId=%s&part=snippet&type=video&maxResults=50&pageToken=%s&order=date",
			apiKey, channelID, pageToken,
		)

		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Failed to fetch videos: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Fatalf("YouTube API error: %s", string(body))
		}

		var searchResp YouTubeSearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
			log.Fatalf("Failed to decode response: %v", err)
		}

		for _, item := range searchResp.Items {
			videoID := item.ID.VideoID
			title := item.Snippet.Title
			publishedStr := item.Snippet.PublishedAt

			published, err := time.Parse(time.RFC3339, publishedStr)
			if err != nil {
				log.Printf("Failed to parse date %s: %v", publishedStr, err)
				continue
			}

			video := Video{
				ID:        videoID,
				Title:     title,
				Published: published,
				URL:       fmt.Sprintf("https://www.youtube.com/embed/%s", videoID),
				Year:      published.Year(),
			}
			videos = append(videos, video)
		}

		if searchResp.NextPageToken == "" {
			break
		}
		pageToken = searchResp.NextPageToken
	}

	// Sort videos by date (newest first)
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].Published.After(videos[j].Published)
	})

	// Generate markdown content
	markdown := generateMarkdown(videos)

	// Write to file
	outputPath := "data/videos.md"
	if err := os.WriteFile(outputPath, []byte(markdown), 0644); err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	fmt.Printf("Successfully fetched %d videos and wrote to %s\n", len(videos), outputPath)
}

func generateMarkdown(videos []Video) string {
	var sb strings.Builder

	sb.WriteString("# Go Meetup Prague Videos\n\n")
	sb.WriteString("Auto-generated list of all videos from the Go Meetup Prague YouTube channel.\n\n")

	// Group videos by year
	videosByYear := make(map[int][]Video)
	for _, video := range videos {
		videosByYear[video.Year] = append(videosByYear[video.Year], video)
	}

	// Sort years in descending order
	var years []int
	for year := range videosByYear {
		years = append(years, year)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(years)))

	// Generate markdown for each year
	for _, year := range years {
		sb.WriteString(fmt.Sprintf("## %d\n\n", year))
		for _, video := range videosByYear[year] {
			sb.WriteString(fmt.Sprintf("### %s\n", video.Title))
			sb.WriteString(fmt.Sprintf("- **Video ID**: %s\n", video.ID))
			sb.WriteString(fmt.Sprintf("- **Published**: %s\n", video.Published.Format("2006-01-02")))
			sb.WriteString(fmt.Sprintf("- **URL**: %s\n\n", video.URL))
		}
	}

	return sb.String()
}

func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Warning: Could not open .env file at %s: %v\n", path, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}
}
