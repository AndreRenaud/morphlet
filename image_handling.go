package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

// drawLine implements Bresenhams line drawing algorithm
func drawLine(img draw.Image, x1, y1, x2, y2 int, col color.Color) {
	dx := x2 - x1
	if dx < 0 {
		dx = -dx
	}
	dy := y2 - y1
	if dy < 0 {
		dy = -dy
	}
	sx := 1
	if x1 >= x2 {
		sx = -1
	}
	sy := 1
	if y1 >= y2 {
		sy = -1
	}
	err := dx - dy
	for {
		img.Set(x1, y1, col)
		if x1 == x2 && y1 == y2 {
			break
		}
		err2 := err * 2
		if err2 > -dy {
			err -= dy
			x1 += sx
		}
		if err2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func drawTriangle(img draw.Image, p1, p2, p3 image.Point, col color.Color) {
	drawLine(img, p1.X, p1.Y, p2.X, p2.Y, col)
	drawLine(img, p2.X, p2.Y, p3.X, p3.Y, col)
	drawLine(img, p3.X, p3.Y, p1.X, p1.Y, col)
}

func loadImage(filename string) (*image.RGBA, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	rgbaImg, ok := img.(*image.RGBA)
	if !ok {
		rgbaImg = image.NewRGBA(img.Bounds())
		draw.Draw(rgbaImg, rgbaImg.Bounds(), img, image.Point{}, draw.Src)
	}

	return rgbaImg, nil
}

func saveImage(img image.Image, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return err
	}
	return nil
}
