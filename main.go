package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math/rand"
	"os"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"github.com/AllenDang/giu"
)

type PointPair struct {
	P1, P2 image.Point
	Color  color.RGBA
}

type TextureWithSize struct {
	Texture *giu.Texture
	Size    image.Point
}

var (
	showProjectView bool
	images          []string
	selectedImage           = -1
	splitSize       float32 = 250
	newImagePath    string
	points          = make(map[int][]PointPair) // key: selectedImage index
	textures        = make(map[string]*TextureWithSize)
	draggingPoint   struct {
		pairSetIndex int  // Corresponds to `pairIndex` in imageWithPoints, which is `selectedImage`
		pairIndex    int  // Index in the `points[pairSetIndex]` slice
		isP1         bool // true if dragging P1, false for P2
		isDragging   bool
	}
)

func onNewProject() {
	showProjectView = true
	images = []string{}
	selectedImage = -1
	points = make(map[int][]PointPair)
	textures = make(map[string]*TextureWithSize)
}

func onOpenProject() {
	onNewProject()
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
		giu.Style().SetFontSize(24),
		giu.Layout{
			giu.Label("MorphLet"),
		},
		giu.Style().SetFontSize(16),
		giu.Layout{
			giu.Button("Start New Project").OnClick(onNewProject),
			giu.Button("Open Project").OnClick(onOpenProject),
		},
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
	imageWidgets := make([]giu.Widget, len(images))
	for i := range images {
		imgLabel := fmt.Sprintf("%d: %s", i+1, images[i])
		localI := i
		imageWidgets[i] = giu.Selectable(imgLabel).Selected(selectedImage == localI).OnClick(func() {
			selectedImage = localI
		})
	}

	return giu.Layout{
		giu.Label("Images"),
		giu.InputText(&newImagePath).Hint("path/to/image.png"),
		giu.Row(
			giu.Button("Add Image").OnClick(func() {
				if newImagePath != "" {
					images = append(images, newImagePath)
					newImagePath = ""
				}
			}),
			giu.Button("Remove Image").OnClick(func() {
				if selectedImage >= 0 && selectedImage < len(images) {
					delete(points, selectedImage)
					delete(points, selectedImage+1)

					images = append(images[:selectedImage], images[selectedImage+1:]...)
					if selectedImage >= len(images) {
						selectedImage = len(images) - 1
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
		availW, availH := giu.GetAvailableRegion()

		if selectedImage == -1 {
			layouts = append(layouts, giu.Label("Please select an image from the list."))
		} else if len(images) > 0 {
			if selectedImage == 0 && len(images) == 1 {
				layouts = append(layouts, giu.Label("This is the only image."))
				_, size, err := loadImage(images[selectedImage])
				if err != nil {
					layouts = append(layouts, giu.Label(err.Error()))
				} else {
					scaledSize := getScaledSize(size, image.Pt(int(availW), int(availH)))
					layouts = append(layouts, giu.Row(imageWithPoints(images[selectedImage], selectedImage, true, scaledSize)))
				}
			} else if selectedImage == 0 && len(images) > 1 {
				layouts = append(layouts, giu.Label("This is the first image. Compare with next ->"))
				_, size, err := loadImage(images[selectedImage])
				if err != nil {
					layouts = append(layouts, giu.Label(err.Error()))
				} else {
					scaledSize := getScaledSize(size, image.Pt(int(availW), int(availH)))
					layouts = append(layouts, giu.Row(imageWithPoints(images[selectedImage], selectedImage+1, true, scaledSize)))
				}
			} else { // selectedImage > 0
				prevImgPath := images[selectedImage-1]
				currImgPath := images[selectedImage]

				_, prevSize, errPrev := loadImage(prevImgPath)
				_, currSize, errCurr := loadImage(currImgPath)

				if errPrev != nil {
					layouts = append(layouts, giu.Label(fmt.Sprintf("Failed to load previous image: %v", errPrev)))
				}
				if errCurr != nil {
					layouts = append(layouts, giu.Label(fmt.Sprintf("Failed to load current image: %v", errCurr)))
				}

				if errPrev == nil && errCurr == nil {
					targetWidth := availW / 2
					prevScaledSize := getScaledSize(prevSize, image.Pt(int(targetWidth), int(availH)))
					currScaledSize := getScaledSize(currSize, image.Pt(int(targetWidth), int(availH)))

					layouts = append(layouts, giu.Row(
						imageWithPoints(prevImgPath, selectedImage, true, prevScaledSize),
						imageWithPoints(currImgPath, selectedImage, false, currScaledSize),
					))
				}
			}
		}

		giu.Layout{
			giu.Label("Image Comparison"),
			giu.Button("Generate Movie").OnClick(func() {
				log.Println("Generate Movie button clicked. Morphing not implemented yet.")
			}),
			giu.Column(layouts...),
		}.Build()
	})
}

func getScaledSize(originalSize, availableSize image.Point) image.Point {
	if originalSize.X == 0 || originalSize.Y == 0 {
		return image.Point{}
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

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, image.Point{}, fmt.Errorf("failed to decode image %s: %w", path, err)
	}

	size := img.Bounds().Size()

	var rgba *image.RGBA
	if i, ok := img.(*image.RGBA); ok {
		rgba = i
	} else {
		rgba = image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)
		img = rgba
	}

	giu.NewTextureFromRgba(img, func(t *giu.Texture) {
		textures[path] = &TextureWithSize{Texture: t, Size: size}
	})
	return nil, image.Point{}, nil
}

func imageWithPoints(imgPath string, pairIndex int, isPrevImage bool, scaledSize image.Point) giu.Widget {
	return giu.Custom(func() {
		tex, originalSize, err := loadImage(imgPath)
		if err != nil {
			giu.Label(err.Error()).Build()
			return
		}

		startPos := giu.GetCursorScreenPos()
		imgWidget := giu.Image(tex).Size(float32(scaledSize.X), float32(scaledSize.Y))
		imgWidget.Build()

		canvas := giu.GetCanvas()

		scaleX := float32(scaledSize.X) / float32(originalSize.X)
		scaleY := float32(scaledSize.Y) / float32(originalSize.Y)

		if draggingPoint.isDragging && giu.IsMouseReleased(giu.MouseButtonLeft) {
			draggingPoint.isDragging = false
		}

		if draggingPoint.isDragging && draggingPoint.pairSetIndex == pairIndex {
			isCorrectPointToDrag := (isPrevImage && draggingPoint.isP1) || (!isPrevImage && !draggingPoint.isP1)
			if isCorrectPointToDrag {
				mousePos := giu.GetMousePos()
				newPos := mousePos.Sub(startPos)
				originalX := int(float32(newPos.X) / scaleX)
				originalY := int(float32(newPos.Y) / scaleY)
				if draggingPoint.isP1 {
					points[pairIndex][draggingPoint.pairIndex].P1 = image.Pt(originalX, originalY)
				} else {
					points[pairIndex][draggingPoint.pairIndex].P2 = image.Pt(originalX, originalY)
				}
			}
		}

		if pointPairs, ok := points[pairIndex]; ok {
			for i, p := range pointPairs {
				var pointOnImage image.Point
				if isPrevImage {
					pointOnImage = p.P1
				} else {
					pointOnImage = p.P2
				}

				scaledX := int(float32(pointOnImage.X) * scaleX)
				scaledY := int(float32(pointOnImage.Y) * scaleY)
				drawPos := startPos.Add(image.Pt(scaledX, scaledY))

				// Flashing logic for the non-dragged point
				isOppositePoint := draggingPoint.isDragging &&
					draggingPoint.pairSetIndex == pairIndex &&
					draggingPoint.pairIndex == i &&
					((isPrevImage && !draggingPoint.isP1) || (!isPrevImage && draggingPoint.isP1))

				if isOppositePoint {
					// Flash every half second
					if time.Now().UnixMilli()/500%2 == 0 {
						canvas.AddCircle(drawPos, 8, color.RGBA{R: 255, G: 255, B: 0, A: 255}, 12, 2)
					}
				}

				canvas.AddCircleFilled(drawPos, 4, p.Color)

				if draggingPoint.isDragging && draggingPoint.pairSetIndex == pairIndex && draggingPoint.pairIndex == i {
					isCorrectPointToDrag := (isPrevImage && draggingPoint.isP1) || (!isPrevImage && !draggingPoint.isP1)
					if isCorrectPointToDrag {
						canvas.AddCircle(drawPos, 8, color.RGBA{R: 255, G: 255, B: 0, A: 255}, 12, 2)
					}
				}
			}
		}

		if giu.IsItemHovered() {
			if giu.IsMouseClicked(giu.MouseButtonLeft) {
				mousePos := giu.GetMousePos()
				clickPos := mousePos.Sub(startPos)

				clickedOnPoint := false
				if pointPairs, ok := points[pairIndex]; ok {
					for i := len(pointPairs) - 1; i >= 0; i-- {
						p := pointPairs[i]
						var pointOnImage image.Point
						if isPrevImage {
							pointOnImage = p.P1
						} else {
							pointOnImage = p.P2
						}

						scaledX := int(float32(pointOnImage.X) * scaleX)
						scaledY := int(float32(pointOnImage.Y) * scaleY)
						distVec := clickPos.Sub(image.Pt(scaledX, scaledY))
						if distVec.X*distVec.X+distVec.Y*distVec.Y < 10*10 {
							draggingPoint.isDragging = true
							draggingPoint.pairSetIndex = pairIndex
							draggingPoint.pairIndex = i
							draggingPoint.isP1 = isPrevImage
							clickedOnPoint = true
							break
						}
					}
				}

				if !clickedOnPoint {
					originalX := int(float32(clickPos.X) / scaleX)
					originalY := int(float32(clickPos.Y) / scaleY)
					newPoint := image.Pt(originalX, originalY)
					newColor := color.RGBA{R: uint8(rand.Intn(256)), G: uint8(rand.Intn(256)), B: uint8(rand.Intn(256)), A: 255}
					if _, ok := points[pairIndex]; !ok {
						points[pairIndex] = []PointPair{}
					}
					points[pairIndex] = append(points[pairIndex], PointPair{P1: newPoint, P2: newPoint, Color: newColor})
				}
			} else if giu.IsMouseClicked(giu.MouseButtonRight) {
				mousePos := giu.GetMousePos()
				clickPos := mousePos.Sub(startPos)

				if pointPairs, ok := points[pairIndex]; ok {
					closestDistSq := 10 * 10
					deleteIndex := -1

					for i, p := range pointPairs {
						var pointOnImage image.Point
						if isPrevImage {
							pointOnImage = p.P1
						} else {
							pointOnImage = p.P2
						}

						scaledX := int(float32(pointOnImage.X) * scaleX)
						scaledY := int(float32(pointOnImage.Y) * scaleY)
						distVec := clickPos.Sub(image.Pt(scaledX, scaledY))
						distSq := distVec.X*distVec.X + distVec.Y*distVec.Y
						if distSq < closestDistSq {
							closestDistSq = distSq
							deleteIndex = i
						}
					}

					if deleteIndex != -1 {
						points[pairIndex] = append(pointPairs[:deleteIndex], pointPairs[deleteIndex+1:]...)
					}
				}
			}
		}
	})
}

func main() {
	wnd := giu.NewMasterWindow("MorphLet", 1024, 768, 0)
	wnd.Run(loop)
}
