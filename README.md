# MorphLet

A Go-based image morphing application with an interactive GUI for creating smooth transitions between images using point correspondence and Delaunay triangulation.

## Features

- **Interactive GUI**: Visual point placement and editing with drag-and-drop functionality
- **Multi-image projects**: Support for morphing sequences with multiple images
- **Point correspondence**: Click to add points, drag to adjust positions, double-click to add points across all images
- **Project management**: Save and load projects as JSON files
- **Image reordering**: Organize image sequences with up/down controls
- **Real-time preview**: Side-by-side image comparison for precise point placement

## Usage

### GUI Mode
```bash
./gui                      # Start with empty project
./gui --job project.json   # Load existing project
```

### Command Line
```bash
# Convert generated images to video
ffmpeg -y -f image2 -framerate 10 -i warped-%05d.png -vcodec libx264 -pix_fmt yuv420p video.mp4
```

## Building

```bash
go build ./cmd/gui
```

## How It Works

1. Load multiple images into a project
2. Place corresponding points on each image by clicking
3. Fine-tune point positions by dragging
4. Generate morphed sequences using Delaunay triangulation
5. Export results as image sequences or video

The morphing algorithm uses Delaunay triangulation to create a mesh of triangles, then applies affine transformations to warp each triangle between corresponding points in the source and destination images.