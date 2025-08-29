package main

import (
	"flag"
	"image"
	"log"

	"github.com/AndreRenaud/morphlet/warp"
	"github.com/fogleman/delaunay"

	_ "image/jpeg"
)

func main() {
	frameCount := flag.Int("frames", 21, "Number of frames to generate")
	flag.Parse()

	imageNames := []string{
		"Newall Timelapse Photos/IMG_20171019_170326.jpg",
		"Newall Timelapse Photos/20171231_142904.jpg",
		"Newall Timelapse Photos/20180309_183059.jpg",
		"Newall Timelapse Photos/20180504_145731.jpg",
		"Newall Timelapse Photos/20190123_121335.jpg",
	}

	images := []*image.NRGBA{}
	for _, imageName := range imageNames {
		img, err := loadImage(imageName)
		if err != nil {
			log.Fatalf("failed to load image %s: %v", imageName, err)
		}
		images = append(images, img)
	}

	job := warp.WarpJob{
		Images: images,
		// Each row is top left, top right, bottom left, bottom right
		ImagePoints: [][]delaunay.Point{
			{{1150, 750}, {2480, 653}, {1061, 1524}, {2413, 1521}},
			{{1551, 841}, {2955, 657}, {1566, 1787}, {3002, 1778}},
			{{1205, 974}, {2573, 793}, {1167, 1916}, {2580, 1966}},
			{{1367, 990}, {2798, 817}, {1321, 1902}, {2779, 2016}},
			{{1026, 1066}, {2430, 919}, {960, 2067}, {2393, 2117}},
		},
	}
	if err := job.Run("warped", *frameCount); err != nil {
		log.Fatalf("failed to run warp job: %v", err)
	}
}
