package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

const youtubeSearchEndpoint = "https://www.googleapis.com/youtube/v3/search"

type Video struct {
	ID        string
	Title     string
	Published time.Time
	URL       string
	Year      int
}

type videoOutput struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Published string `json:"published"`
	URL       string `json:"url"`
}

type yearOutput struct {
	Year   int           `json:"year"`
	Videos []videoOutput `json:"videos"`
}

type videosData struct {
	Years []yearOutput `json:"years"`
}

type youtubeSearchResponse struct {
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
	loadEnvFile("../.env")

	apiKey := os.Getenv("GOOGLE_API_KEY")
	channelID := os.Getenv("CHANNEL_ID")
	if apiKey == "" || channelID == "" {
		log.Fatal("GOOGLE_API_KEY and CHANNEL_ID must be set in the environment or ../.env")
	}

	client := &http.Client{Timeout: 20 * time.Second}
	videos, err := fetchVideos(context.Background(), client, youtubeSearchEndpoint, apiKey, channelID)
	if err != nil {
		log.Fatalf("Failed to fetch videos: %v", err)
	}

	outputPath := "../data/videos.json"
	if len(os.Args) > 1 {
		outputPath = os.Args[1]
	}
	if err := writeVideos(outputPath, videos); err != nil {
		log.Fatalf("Failed to write videos: %v", err)
	}

	fmt.Printf("Successfully fetched %d videos and wrote to %s\n", len(videos), outputPath)
}

func fetchVideos(ctx context.Context, client *http.Client, endpoint, apiKey, channelID string) ([]Video, error) {
	var videos []Video
	pageToken := ""

	for {
		response, err := fetchPage(ctx, client, endpoint, apiKey, channelID, pageToken)
		if err != nil {
			return nil, err
		}

		for _, item := range response.Items {
			published, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
			if err != nil {
				return nil, fmt.Errorf("parse published date %q: %w", item.Snippet.PublishedAt, err)
			}
			if item.ID.VideoID == "" {
				continue
			}
			videos = append(videos, Video{
				ID:        item.ID.VideoID,
				Title:     html.UnescapeString(item.Snippet.Title),
				Published: published,
				URL:       "https://www.youtube.com/embed/" + item.ID.VideoID,
				Year:      published.Year(),
			})
		}

		if response.NextPageToken == "" {
			break
		}
		pageToken = response.NextPageToken
	}

	sort.Slice(videos, func(i, j int) bool {
		return videos[i].Published.After(videos[j].Published)
	})
	return videos, nil
}

func fetchPage(ctx context.Context, client *http.Client, endpoint, apiKey, channelID, pageToken string) (youtubeSearchResponse, error) {
	requestURL, err := url.Parse(endpoint)
	if err != nil {
		return youtubeSearchResponse{}, fmt.Errorf("parse YouTube endpoint: %w", err)
	}
	query := requestURL.Query()
	query.Set("key", apiKey)
	query.Set("channelId", channelID)
	query.Set("part", "snippet")
	query.Set("type", "video")
	query.Set("maxResults", "50")
	query.Set("order", "date")
	if pageToken != "" {
		query.Set("pageToken", pageToken)
	}
	requestURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return youtubeSearchResponse{}, fmt.Errorf("create YouTube request: %w", err)
	}
	response, err := client.Do(request)
	if err != nil {
		return youtubeSearchResponse{}, fmt.Errorf("request YouTube API: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(io.LimitReader(response.Body, 4096))
		if readErr != nil {
			return youtubeSearchResponse{}, fmt.Errorf("YouTube API returned %s", response.Status)
		}
		return youtubeSearchResponse{}, fmt.Errorf("YouTube API returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}

	var result youtubeSearchResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return youtubeSearchResponse{}, fmt.Errorf("decode YouTube response: %w", err)
	}
	return result, nil
}

func writeVideos(path string, videos []Video) error {
	years := make(map[int][]videoOutput)
	for _, video := range videos {
		years[video.Year] = append(years[video.Year], videoOutput{
			ID:        video.ID,
			Title:     video.Title,
			Published: video.Published.Format("2006-01-02"),
			URL:       video.URL,
		})
	}

	yearNumbers := make([]int, 0, len(years))
	for year := range years {
		yearNumbers = append(yearNumbers, year)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(yearNumbers)))

	data := videosData{Years: make([]yearOutput, 0, len(yearNumbers))}
	for _, year := range yearNumbers {
		data.Years = append(data.Years, yearOutput{Year: year, Videos: years[year]})
	}

	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encode videos: %w", err)
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(path, encoded, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("Warning: could not open %s: %v", path, err)
		}
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if _, exists := os.LookupEnv(key); !exists && value != "" {
			if err := os.Setenv(key, value); err != nil {
				log.Printf("Warning: could not set %s: %v", key, err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Warning: could not read %s: %v", path, err)
	}
}
