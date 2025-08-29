package main

import (
	"fmt"
	"image"
	"image/draw"
	"os"

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
	pendingPoint    *image.Point // Store a point from image 0 waiting to be paired
)

func onNewProject() {
	showProjectView = true
	currentJob = &warp.WarpJobSaveFormat{
		Images:      []string{},
		ImagePoints: [][][]int{},
	}
	pendingPoint = nil // Clear any pending point
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
	)
}

func projectUI() {
	giu.SingleWindow().Layout(
		giu.SplitLayout(giu.DirectionHorizontal, &splitSize,
			imagePane(),
			comparisonPane(),
		),
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
		imageWidgets[i] = giu.Selectable(imgLabel).Selected(selectedImage == localI).OnClick(func() {
			selectedImage = localI
			pendingPoint = nil // Clear pending point when switching images
		})
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
					pendingPoint = nil // Clear pending point
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

				// Show pending point status
				if pendingPoint != nil {
					layouts = append(layouts, giu.Label(fmt.Sprintf("Pending point on image 0: (%d, %d) - Click on image %d to pair", pendingPoint.X, pendingPoint.Y, selectedImage)))
				} else {
					layouts = append(layouts, giu.Label("Click on image 0 first, then on the corresponding point on the other image"))
				}

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
							clickableImage(tex0, scaledSize0, size0, true),
						),
						giu.Column(
							giu.Label(fmt.Sprintf("Image %d", selectedImage)),
							clickableImage(texSelected, scaledSizeSelected, sizeSelected, false),
						),
					))
				}
			}
		}

		giu.Layout{
			giu.Label("Image Comparison"),
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

func addPointPair(image0Point, image1Point image.Point) {
	if currentJob == nil || selectedImage <= 0 {
		return
	}

	// Ensure we have enough space in ImagePoints
	pairIndex := selectedImage - 1 // Use selectedImage-1 as the pair index
	for len(currentJob.ImagePoints) <= pairIndex {
		currentJob.ImagePoints = append(currentJob.ImagePoints, [][]int{})
	}

	// Add the point pair [x0, y0, x1, y1] where 0 is first image, 1 is second image
	pointPair := []int{image0Point.X, image0Point.Y, image1Point.X, image1Point.Y}
	currentJob.ImagePoints[pairIndex] = append(currentJob.ImagePoints[pairIndex], pointPair)
}

func clickableImage(tex *giu.Texture, scaledSize image.Point, originalSize image.Point, isImage0 bool) giu.Widget {
	return giu.Custom(func() {
		startPos := giu.GetCursorScreenPos()
		imgWidget := giu.Image(tex).Size(float32(scaledSize.X), float32(scaledSize.Y))
		imgWidget.Build()

		// Handle mouse clicks on the image
		if giu.IsItemHovered() && giu.IsMouseClicked(giu.MouseButtonLeft) {
			mousePos := giu.GetMousePos()
			clickPos := mousePos.Sub(startPos)

			// Convert click position to original image coordinates
			scaleX := float32(originalSize.X) / float32(scaledSize.X)
			scaleY := float32(originalSize.Y) / float32(scaledSize.Y)
			originalX := int(float32(clickPos.X) * scaleX)
			originalY := int(float32(clickPos.Y) * scaleY)

			// Only add points when we have two images being compared
			if selectedImage > 0 {
				if isImage0 {
					// Store the click on image 0 as pending
					point := image.Pt(originalX, originalY)
					pendingPoint = &point
				} else {
					// This is a click on the selected image (image 1)
					if pendingPoint != nil {
						// We have a pending point, create the pair
						addPointPair(*pendingPoint, image.Pt(originalX, originalY))
						pendingPoint = nil // Clear the pending point
					} else {
						// No pending point, just store this as pending
						point := image.Pt(originalX, originalY)
						pendingPoint = &point
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
	wnd := giu.NewMasterWindow("MorphLet", 1024, 768, 0)
	wnd.SetStyle(Theme())
	wnd.Run(loop)
}
