package warp

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

func LoadImage(filename string) (*image.NRGBA, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	rgbaImg, ok := img.(*image.NRGBA)
	if !ok {
		rgbaImg = image.NewNRGBA(img.Bounds())
		draw.Draw(rgbaImg, rgbaImg.Bounds(), img, image.Point{}, draw.Src)
	}

	return rgbaImg, nil
}

func SaveImage(img image.Image, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	ext := filepath.Ext(filename)
	switch ext {
	case ".jpeg", ".jpg":
		if err := jpeg.Encode(file, img, nil); err != nil {
			return err
		}
	case ".png":
		if err := png.Encode(file, img); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported image format: %s", ext)
	}

	return nil
}
