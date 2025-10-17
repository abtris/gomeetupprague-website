#!/usr/bin/env python3
"""
Fetch all videos from Go Meetup Prague YouTube channel using YouTube API v3
and generate a markdown file with video details.
"""

import os
import json
import sys
from datetime import datetime
from pathlib import Path
from urllib.request import urlopen
from urllib.parse import urlencode


def load_env_file(env_path):
    """Load environment variables from .env file."""
    if not env_path.exists():
        return

    with open(env_path) as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            if "=" in line:
                key, value = line.split("=", 1)
                os.environ[key.strip()] = value.strip()


# Load environment variables from .env file
env_file = Path(__file__).parent.parent / ".env"
load_env_file(env_file)

GOOGLE_API_KEY = os.getenv("GOOGLE_API_KEY")
CHANNEL_ID = os.getenv("CHANNEL_ID")

if not GOOGLE_API_KEY:
    print("Error: GOOGLE_API_KEY not set in .env file")
    sys.exit(1)

if not CHANNEL_ID:
    print("Error: CHANNEL_ID not set in .env file")
    sys.exit(1)


def fetch_videos():
    """Fetch all videos from the YouTube channel."""
    videos = []
    page_token = None
    
    while True:
        params = {
            "key": GOOGLE_API_KEY,
            "channelId": CHANNEL_ID,
            "part": "snippet",
            "type": "video",
            "maxResults": 50,
            "order": "date",
        }
        
        if page_token:
            params["pageToken"] = page_token
        
        url = f"https://www.googleapis.com/youtube/v3/search?{urlencode(params)}"
        
        try:
            with urlopen(url) as response:
                data = json.loads(response.read().decode())
        except Exception as e:
            print(f"Error fetching videos: {e}")
            sys.exit(1)
        
        if "error" in data:
            print(f"YouTube API error: {data['error']}")
            sys.exit(1)
        
        for item in data.get("items", []):
            video_id = item["id"]["videoId"]
            title = item["snippet"]["title"]
            published_at = item["snippet"]["publishedAt"]
            
            # Parse the published date
            published = datetime.fromisoformat(published_at.replace("Z", "+00:00"))
            
            videos.append({
                "id": video_id,
                "title": title,
                "published": published,
                "url": f"https://www.youtube.com/embed/{video_id}",
                "year": published.year,
            })
        
        page_token = data.get("nextPageToken")
        if not page_token:
            break
    
    return videos


def generate_markdown(videos):
    """Generate markdown content from videos."""
    # Sort videos by date (newest first)
    videos.sort(key=lambda v: v["published"], reverse=True)
    
    # Group videos by year
    videos_by_year = {}
    for video in videos:
        year = video["year"]
        if year not in videos_by_year:
            videos_by_year[year] = []
        videos_by_year[year].append(video)
    
    # Generate markdown
    lines = [
        "# Go Meetup Prague Videos",
        "",
        "Auto-generated list of all videos from the Go Meetup Prague YouTube channel.",
        "",
    ]
    
    # Sort years in descending order
    for year in sorted(videos_by_year.keys(), reverse=True):
        lines.append(f"## {year}")
        lines.append("")
        
        for video in videos_by_year[year]:
            lines.append(f"### {video['title']}")
            lines.append(f"- **Video ID**: {video['id']}")
            lines.append(f"- **Published**: {video['published'].strftime('%Y-%m-%d')}")
            lines.append(f"- **URL**: {video['url']}")
            lines.append("")
    
    return "\n".join(lines)


def main():
    """Main function."""
    print("Fetching videos from Go Meetup Prague YouTube channel...")
    videos = fetch_videos()
    print(f"Found {len(videos)} videos")
    
    markdown = generate_markdown(videos)
    
    output_path = Path(__file__).parent.parent / "data" / "videos.md"
    output_path.parent.mkdir(parents=True, exist_ok=True)
    
    with open(output_path, "w") as f:
        f.write(markdown)
    
    print(f"Successfully wrote {len(videos)} videos to {output_path}")


if __name__ == "__main__":
    main()

