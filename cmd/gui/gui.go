package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"github.com/AllenDang/giu"
	"github.com/AndreRenaud/morphlet/warp"
)

type TextureWithSize struct {
	Texture *giu.Texture
	Size    image.Point
}

var (
	showProjectView bool
	currentJob      *warp.WarpJobSaveFormat
	selectedImage           = -1
	splitSize       float32 = 250
	newImagePath    string
	textures        = make(map[string]*TextureWithSize)
	draggingPoint   struct {
		isDragging   bool
		imageIndex   int
		pointIndex   int
		dragStartPos image.Point
	}
	lastClickTime float64
	lastClickPos  image.Point
	saveFilePath  string
)

func onNewProject() {
	showProjectView = true
	currentJob = &warp.WarpJobSaveFormat{
		Images:      []string{},
		ImagePoints: [][][]int{},
	}
}

func onOpenProject() {
	// TODO: implement loading
}

func loop() {
	if showProjectView {
		projectUI()
	} else {
		startUI()
	}
}

func startUI() {
	giu.SingleWindow().Layout(
		giu.Label("Welcome to MorphLet"),
		giu.Button("New Project").OnClick(onNewProject),
		giu.Button("Open Project").OnClick(onOpenProject),
		giu.PrepareMsgbox(),
	)
}

func projectUI() {
	giu.SingleWindow().Layout(
		giu.SplitLayout(giu.DirectionHorizontal, &splitSize,
			imagePane(),
			comparisonPane(),
		),
		giu.PrepareMsgbox(),
	)
}

func imagePane() giu.Widget {
	if currentJob == nil {
		return giu.Layout{
			giu.Label("No project loaded"),
		}
	}

	imageWidgets := make([]giu.Widget, len(currentJob.Images))
	for i := range currentJob.Images {
		imgLabel := fmt.Sprintf("%d: %s", i+1, currentJob.Images[i])
		localI := i

		// Create a row with selectable image name and up/down buttons on the right
		imageWidgets[i] = giu.Row(
			giu.Selectable(imgLabel).Selected(selectedImage == localI).OnClick(func() {
				selectedImage = localI
			}), // Let the selectable take available space
			giu.Dummy(1, 0), // Spacer to push buttons to the right
			giu.Button("↑").Size(25, 0).OnClick(func() {
				// Move image up (swap with previous)
				if localI > 0 && currentJob != nil {
					// Swap images
					currentJob.Images[localI], currentJob.Images[localI-1] = currentJob.Images[localI-1], currentJob.Images[localI]

					// Swap image points if they exist
					if len(currentJob.ImagePoints) > localI && len(currentJob.ImagePoints) > localI-1 {
						currentJob.ImagePoints[localI], currentJob.ImagePoints[localI-1] = currentJob.ImagePoints[localI-1], currentJob.ImagePoints[localI]
					}

					// Update selected image index if needed
					if selectedImage == localI {
						selectedImage = localI - 1
					} else if selectedImage == localI-1 {
						selectedImage = localI
					}
				}
			}).Disabled(localI == 0), // Disable up button for first item
			giu.Button("↓").Size(25, 0).OnClick(func() {
				// Move image down (swap with next)
				if localI < len(currentJob.Images)-1 && currentJob != nil {
					// Swap images
					currentJob.Images[localI], currentJob.Images[localI+1] = currentJob.Images[localI+1], currentJob.Images[localI]

					// Swap image points if they exist
					if len(currentJob.ImagePoints) > localI && len(currentJob.ImagePoints) > localI+1 {
						currentJob.ImagePoints[localI], currentJob.ImagePoints[localI+1] = currentJob.ImagePoints[localI+1], currentJob.ImagePoints[localI]
					}

					// Update selected image index if needed
					if selectedImage == localI {
						selectedImage = localI + 1
					} else if selectedImage == localI+1 {
						selectedImage = localI
					}
				}
			}).Disabled(localI == len(currentJob.Images)-1), // Disable down button for last item
		)
	}

	return giu.Layout{
		giu.Label("Images"),
		giu.InputText(&newImagePath).Hint("path/to/image.png"),
		giu.Row(
			giu.Button("Add Image").OnClick(func() {
				if newImagePath != "" {
					currentJob.Images = append(currentJob.Images, newImagePath)
					// Initialize empty point list for new image
					currentJob.ImagePoints = append(currentJob.ImagePoints, [][]int{})
					newImagePath = ""
				}
			}),
			giu.Button("Remove Image").OnClick(func() {
				if selectedImage >= 0 && selectedImage < len(currentJob.Images) {
					// Remove the image and its points
					currentJob.Images = append(currentJob.Images[:selectedImage], currentJob.Images[selectedImage+1:]...)
					currentJob.ImagePoints = append(currentJob.ImagePoints[:selectedImage], currentJob.ImagePoints[selectedImage+1:]...)
					if selectedImage >= len(currentJob.Images) {
						selectedImage = len(currentJob.Images) - 1
					}
				}
			}),
		),
		giu.Column(imageWidgets...),
	}
}

func comparisonPane() giu.Widget {
	return giu.Custom(func() {
		var layouts []giu.Widget

		if currentJob == nil {
			layouts = append(layouts, giu.Label("No project loaded"))
		} else if selectedImage == -1 {
			layouts = append(layouts, giu.Label("Please select an image from the list."))
		} else if len(currentJob.Images) > 0 {
			if selectedImage == 0 && len(currentJob.Images) == 1 {
				// Single image case - just show image 0
				layouts = append(layouts, giu.Label("This is the only image."))
				tex, size, err := loadImage(currentJob.Images[selectedImage])
				if err != nil {
					layouts = append(layouts, giu.Label(err.Error()))
				} else {
					availW, availH := giu.GetAvailableRegion()
					scaledSize := getScaledSize(size, image.Pt(int(availW), int(availH)))
					layouts = append(layouts, giu.Row(simpleImage(tex, scaledSize)))
				}
			} else if selectedImage == 0 && len(currentJob.Images) > 1 {
				// Image 0 selected with multiple images - just show image 0
				layouts = append(layouts, giu.Label("First image selected."))
				tex, size, err := loadImage(currentJob.Images[0])
				if err != nil {
					layouts = append(layouts, giu.Label(err.Error()))
				} else {
					availW, availH := giu.GetAvailableRegion()
					scaledSize := getScaledSize(size, image.Pt(int(availW), int(availH)))
					layouts = append(layouts, giu.Row(simpleImage(tex, scaledSize)))
				}
			} else if selectedImage > 0 {
				// Non-zero image selected - show image 0 beside selected image
				layouts = append(layouts, giu.Label(fmt.Sprintf("Comparing image 0 with image %d", selectedImage)))

				tex0, size0, err0 := loadImage(currentJob.Images[0])
				texSelected, sizeSelected, errSelected := loadImage(currentJob.Images[selectedImage])

				if err0 != nil {
					layouts = append(layouts, giu.Label("Error loading image 0: "+err0.Error()))
				} else if errSelected != nil {
					layouts = append(layouts, giu.Label("Error loading selected image: "+errSelected.Error()))
				} else {
					availW, availH := giu.GetAvailableRegion()

					// Calculate side-by-side layout - each image gets half the width
					halfWidth := int(availW / 2)

					// Scale both images to fit their half of the available space
					scaledSize0 := getScaledSize(size0, image.Pt(halfWidth, int(availH)))
					scaledSizeSelected := getScaledSize(sizeSelected, image.Pt(halfWidth, int(availH)))

					layouts = append(layouts, giu.Row(
						giu.Column(
							giu.Label("Image 0"),
							clickableImage(tex0, scaledSize0, size0, 0),
						),
						giu.Column(
							giu.Label(fmt.Sprintf("Image %d", selectedImage)),
							clickableImage(texSelected, scaledSizeSelected, sizeSelected, selectedImage),
						),
					))
				}
			}
		}

		giu.Layout{
			giu.Label("Image Comparison"),
			giu.Row(
				giu.Button("Generate Morph").OnClick(func() {
					if currentJob != nil && selectedImage > 0 {
						pairIndex := selectedImage - 1
						if len(currentJob.ImagePoints) > pairIndex && len(currentJob.ImagePoints[pairIndex]) > 0 {
							giu.Msgbox("Info", fmt.Sprintf("Would morph with %d point pairs", len(currentJob.ImagePoints[pairIndex])))
						} else {
							giu.Msgbox("Info", "No point pairs available for morphing")
						}
					} else {
						giu.Msgbox("Info", "Select a non-zero image to enable morphing")
					}
				}),
				giu.Button("Save Project").OnClick(func() {
					if currentJob != nil {
						if saveFilePath == "" {
							saveFilePath = "project.json"
						}
						err := warp.SaveWarpJson(currentJob, saveFilePath)
						if err != nil {
							giu.Msgbox("Error", fmt.Sprintf("Failed to save project: %v", err))
						} else {
							giu.Msgbox("Success", fmt.Sprintf("Project saved to: %s", saveFilePath))
						}
					} else {
						giu.Msgbox("Error", "No project to save")
					}
				}),
			),
			giu.InputText(&saveFilePath).Hint("project.json").Label("Save as:"),
			giu.Column(layouts...),
		}.Build()
	})
}

func getScaledSize(originalSize, availableSize image.Point) image.Point {
	if originalSize.X == 0 || originalSize.Y == 0 {
		return image.Point{X: 100, Y: 100}
	}

	ratio := float32(originalSize.X) / float32(originalSize.Y)
	newWidth := float32(availableSize.X)
	newHeight := newWidth / ratio

	if newHeight > float32(availableSize.Y) {
		newHeight = float32(availableSize.Y)
		newWidth = newHeight * ratio
	}

	return image.Point{X: int(newWidth), Y: int(newHeight)}
}

func loadImage(path string) (*giu.Texture, image.Point, error) {
	if tex, ok := textures[path]; ok {
		return tex.Texture, tex.Size, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, image.Point{}, fmt.Errorf("failed to open image %s: %w", path, err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, image.Point{}, fmt.Errorf("failed to decode image %s: %w", path, err)
	}

	bounds := img.Bounds()
	size := bounds.Size()

	// Convert to RGBA
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	var tex *giu.Texture
	giu.NewTextureFromRgba(rgba, func(t *giu.Texture) {
		tex = t
	})
	if tex == nil {
		return nil, image.Point{}, fmt.Errorf("failed to create texture from %s", path)
	}

	textures[path] = &TextureWithSize{Texture: tex, Size: size}
	return tex, size, nil
}

func addPointToAllImages(clickedPoint image.Point, clickedImageIndex int) {
	if currentJob == nil {
		return
	}

	// Ensure we have enough images and space in ImagePoints
	numImages := len(currentJob.Images)
	if numImages == 0 {
		return
	}

	// Ensure ImagePoints has enough entries for all images
	for len(currentJob.ImagePoints) < numImages {
		currentJob.ImagePoints = append(currentJob.ImagePoints, [][]int{})
	}

	// Add a new point to each image at the same relative position
	for i := 0; i < numImages; i++ {
		var pointX, pointY int
		if i == clickedImageIndex {
			// Use the exact clicked position for the clicked image
			pointX = clickedPoint.X
			pointY = clickedPoint.Y
		} else {
			// For other images, use the same coordinates
			// In a more sophisticated implementation, you might want to transform
			// coordinates based on image differences, but for now use the same coords
			pointX = clickedPoint.X
			pointY = clickedPoint.Y
		}

		// Add the point as [x, y] to this image
		newPoint := []int{pointX, pointY}
		currentJob.ImagePoints[i] = append(currentJob.ImagePoints[i], newPoint)
	}
}

func clickableImage(tex *giu.Texture, scaledSize image.Point, originalSize image.Point, imageIndex int) giu.Widget {
	return giu.Custom(func() {
		startPos := giu.GetCursorScreenPos()
		imgWidget := giu.Image(tex).Size(float32(scaledSize.X), float32(scaledSize.Y))
		imgWidget.Build()

		// Get canvas for drawing points
		canvas := giu.GetCanvas()

		// Calculate scale factors for converting original coordinates to display coordinates
		scaleX := float32(scaledSize.X) / float32(originalSize.X)
		scaleY := float32(scaledSize.Y) / float32(originalSize.Y)

		// Handle mouse dragging
		if giu.IsItemHovered() {
			mousePos := giu.GetMousePos()
			clickPos := mousePos.Sub(startPos)

			if giu.IsMouseClicked(giu.MouseButtonLeft) {
				currentTime := time.Now().UnixNano() / int64(time.Millisecond)

				// Check for double-click (within 500ms and 5 pixels)
				timeDiff := float64(currentTime) - lastClickTime
				dx := clickPos.X - lastClickPos.X
				dy := clickPos.Y - lastClickPos.Y
				distSq := dx*dx + dy*dy

				if timeDiff < 500 && distSq <= 5*5 {
					// Double-click detected - add new point to all images
					// Convert to original image coordinates
					originalX := int(float32(clickPos.X) / scaleX)
					originalY := int(float32(clickPos.Y) / scaleY)

					// Clamp to image bounds
					if originalX < 0 {
						originalX = 0
					}
					if originalX >= originalSize.X {
						originalX = originalSize.X - 1
					}
					if originalY < 0 {
						originalY = 0
					}
					if originalY >= originalSize.Y {
						originalY = originalSize.Y - 1
					}

					clickedPoint := image.Pt(originalX, originalY)
					addPointToAllImages(clickedPoint, imageIndex)

					// Reset click tracking to prevent triple-click issues
					lastClickTime = 0
					lastClickPos = image.Pt(0, 0)
				} else {
					// Single click - check if clicking near an existing point for dragging
					if currentJob != nil && len(currentJob.ImagePoints) > imageIndex {
						for pointIdx, pointPair := range currentJob.ImagePoints[imageIndex] {
							if len(pointPair) >= 2 {
								pointOnImage := image.Pt(pointPair[0], pointPair[1])
								displayX := int(float32(pointOnImage.X) * scaleX)
								displayY := int(float32(pointOnImage.Y) * scaleY)
								displayPoint := image.Pt(displayX, displayY)

								// Check if click is within 10 pixels of the point
								dx := clickPos.X - displayPoint.X
								dy := clickPos.Y - displayPoint.Y
								distSq := dx*dx + dy*dy
								if distSq <= 10*10 {
									// Start dragging this point
									draggingPoint.isDragging = true
									draggingPoint.imageIndex = imageIndex
									draggingPoint.pointIndex = pointIdx
									draggingPoint.dragStartPos = clickPos
									break
								}
							}
						}
					}

					// Update click tracking for double-click detection
					lastClickTime = float64(currentTime)
					lastClickPos = clickPos
				}
			}
		}

		// Handle dragging update
		if draggingPoint.isDragging && draggingPoint.imageIndex == imageIndex {
			if giu.IsMouseDown(giu.MouseButtonLeft) {
				mousePos := giu.GetMousePos()
				currentPos := mousePos.Sub(startPos)

				// Convert to original image coordinates
				originalX := int(float32(currentPos.X) / scaleX)
				originalY := int(float32(currentPos.Y) / scaleY)

				// Clamp to image bounds
				if originalX < 0 {
					originalX = 0
				}
				if originalX >= originalSize.X {
					originalX = originalSize.X - 1
				}
				if originalY < 0 {
					originalY = 0
				}
				if originalY >= originalSize.Y {
					originalY = originalSize.Y - 1
				}

				// Update the point position
				if currentJob != nil && len(currentJob.ImagePoints) > imageIndex &&
					draggingPoint.pointIndex < len(currentJob.ImagePoints[imageIndex]) {
					currentJob.ImagePoints[imageIndex][draggingPoint.pointIndex][0] = originalX
					currentJob.ImagePoints[imageIndex][draggingPoint.pointIndex][1] = originalY
				}
			} else {
				// Mouse released, stop dragging
				draggingPoint.isDragging = false
			}
		}

		// Draw existing points
		if currentJob != nil && len(currentJob.ImagePoints) > imageIndex {
			for i, pointPair := range currentJob.ImagePoints[imageIndex] {
				if len(pointPair) >= 2 {
					pointOnImage := image.Pt(pointPair[0], pointPair[1])

					// Convert to display coordinates
					displayX := int(float32(pointOnImage.X) * scaleX)
					displayY := int(float32(pointOnImage.Y) * scaleY)
					drawPos := startPos.Add(image.Pt(displayX, displayY))

					// Draw the point with a unique color for each pair
					colors := []color.RGBA{
						{R: 255, G: 0, B: 0, A: 255},   // Red
						{R: 0, G: 255, B: 0, A: 255},   // Green
						{R: 0, G: 0, B: 255, A: 255},   // Blue
						{R: 255, G: 255, B: 0, A: 255}, // Yellow
						{R: 255, G: 0, B: 255, A: 255}, // Magenta
						{R: 0, G: 255, B: 255, A: 255}, // Cyan
					}
					pointColor := colors[i%len(colors)]

					// Highlight the point being dragged
					if draggingPoint.isDragging && draggingPoint.imageIndex == imageIndex && draggingPoint.pointIndex == i {
						canvas.AddCircleFilled(drawPos, 6, pointColor)
						canvas.AddCircle(drawPos, 8, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 12, 2)
					} else {
						canvas.AddCircleFilled(drawPos, 4, pointColor)
						canvas.AddCircle(drawPos, 6, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 12, 1)
					}
				}
			}
		}
	})
}

func simpleImage(tex *giu.Texture, scaledSize image.Point) giu.Widget {
	return giu.Image(tex).Size(float32(scaledSize.X), float32(scaledSize.Y))
}

func main() {
	// Parse command line arguments
	jobFile := flag.String("job", "", "Load project from JSON file")
	flag.Parse()

	// Load job file if specified
	if *jobFile != "" {
		loadedJob, err := warp.LoadWarpJson(*jobFile)
		if err != nil {
			fmt.Printf("Error loading job file '%s': %v\n", *jobFile, err)
			os.Exit(1)
		}
		// Initialize the project with the loaded job
		currentJob = loadedJob
		showProjectView = true
		// If there are images, select the first one by default
		if len(currentJob.Images) > 0 {
			selectedImage = 0
		}
	}

	wnd := giu.NewMasterWindow("MorphLet", 1024, 768, 0)
	wnd.SetStyle(Theme())
	wnd.Run(loop)
}
