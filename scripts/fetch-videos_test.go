package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFetchVideosHandlesPaginationAndSortsNewestFirst(t *testing.T) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("key") != "test-key" || request.URL.Query().Get("channelId") != "test-channel" {
			t.Error("API credentials were not sent as query parameters")
		}
		writer.Header().Set("Content-Type", "application/json")
		if request.URL.Query().Get("pageToken") == "next" {
			_, _ = writer.Write([]byte(`{"items":[{"id":{"videoId":"new"},"snippet":{"title":"Go &amp; Prague","publishedAt":"2026-08-01T12:00:00Z"}}]}`))
			return
		}
		_, _ = writer.Write([]byte(`{"nextPageToken":"next","items":[{"id":{"videoId":"old"},"snippet":{"title":"Older talk","publishedAt":"2025-02-01T12:00:00Z"}}]}`))
	}))
	defer server.Close()

	videos, err := fetchVideos(context.Background(), server.Client(), server.URL, "test-key", "test-channel")
	if err != nil {
		t.Fatalf("fetchVideos returned an error: %v", err)
	}
	if len(videos) != 2 {
		t.Fatalf("got %d videos, want 2", len(videos))
	}
	if videos[0].ID != "new" || videos[0].Title != "Go & Prague" {
		t.Fatalf("newest video was not sorted and decoded correctly: %#v", videos[0])
	}
}

func TestWriteVideosProducesHugoDataShape(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "videos.json")
	videos := []Video{
		{ID: "new", Title: "New", Published: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), URL: "https://www.youtube.com/embed/new", Year: 2026},
		{ID: "old", Title: "Old", Published: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC), URL: "https://www.youtube.com/embed/old", Year: 2025},
	}

	if err := writeVideos(outputPath, videos); err != nil {
		t.Fatalf("writeVideos returned an error: %v", err)
	}
	contents, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	var data videosData
	if err := json.Unmarshal(contents, &data); err != nil {
		t.Fatalf("generated invalid JSON: %v", err)
	}
	if len(data.Years) != 2 || data.Years[0].Year != 2026 {
		t.Fatalf("years were not grouped newest-first: %#v", data.Years)
	}
	if data.Years[0].Videos[0].Published != "2026-08-01" {
		t.Fatalf("published date has wrong format: %q", data.Years[0].Videos[0].Published)
	}
}

func TestLoadEnvFileDoesNotOverrideWorkflowSecrets(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "workflow-secret")
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("GOOGLE_API_KEY=\nCHANNEL_ID=local-channel\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	loadEnvFile(envPath)
	if got := os.Getenv("GOOGLE_API_KEY"); got != "workflow-secret" {
		t.Fatalf("environment secret was overwritten: %q", got)
	}
	if got := os.Getenv("CHANNEL_ID"); got != "local-channel" {
		t.Fatalf("CHANNEL_ID was not loaded: %q", got)
	}
}
