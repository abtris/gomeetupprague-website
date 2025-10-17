#!/usr/bin/env python3
"""
Convert videos.md to videos.json for Hugo data files.
"""

import json
import re
from pathlib import Path

def parse_videos_md(md_content):
    """Parse videos.md and return structured data."""
    videos_by_year = {}
    current_year = None
    current_video = None
    
    for line in md_content.split('\n'):
        line = line.strip()
        
        if line.startswith('## '):
            current_year = line.replace('## ', '')
            videos_by_year[current_year] = []
        elif line.startswith('### '):
            if current_video:
                videos_by_year[current_year].append(current_video)
            current_video = {
                'title': line.replace('### ', ''),
                'id': '',
                'published': '',
                'url': ''
            }
        elif line.startswith('- **Video ID**: '):
            if current_video:
                current_video['id'] = line.replace('- **Video ID**: ', '')
        elif line.startswith('- **Published**: '):
            if current_video:
                current_video['published'] = line.replace('- **Published**: ', '')
        elif line.startswith('- **URL**: '):
            if current_video:
                current_video['url'] = line.replace('- **URL**: ', '')
    
    if current_video:
        videos_by_year[current_year].append(current_video)
    
    return videos_by_year

def main():
    md_path = Path(__file__).parent.parent / "data" / "videos.md"
    json_path = Path(__file__).parent.parent / "data" / "videos.json"
    
    with open(md_path) as f:
        md_content = f.read()
    
    videos_by_year = parse_videos_md(md_content)
    
    # Convert to list of years with videos
    data = {
        'years': []
    }
    
    for year in sorted(videos_by_year.keys(), reverse=True, key=int):
        data['years'].append({
            'year': int(year),
            'videos': videos_by_year[year]
        })
    
    with open(json_path, 'w') as f:
        json.dump(data, f, indent=2)
    
    print(f"Successfully converted videos.md to {json_path}")

if __name__ == "__main__":
    main()

