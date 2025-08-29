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

	image1Name := "Newall Timelapse Photos/20180309_183059.jpg"
	image2Name := "Newall Timelapse Photos/20190123_121335.jpg"

	img1, err := loadImage(image1Name)
	if err != nil {
		log.Fatalf("failed to load image %s: %v", image1Name, err)
	}
	img2, err := loadImage(image2Name)
	if err != nil {
		log.Fatalf("failed to load image %s: %v", image2Name, err)
	}
	log.Printf("img1: %v", img1.Bounds())
	log.Printf("img2: %v", img2.Bounds())
	job := warp.WarpJob{
		Images: []*image.NRGBA{img1, img2},
		// Each row is top left, top right, bottom left, bottom right
		ImagePoints: [][]delaunay.Point{
			{{1205, 974}, {2573, 793}, {1167, 1916}, {2580, 1966}},
			{{1026, 1066}, {2430, 919}, {960, 2067}, {2393, 2117}},
		},
	}
	if err := job.Run("warped", *frameCount); err != nil {
		log.Fatalf("failed to run warp job: %v", err)
	}
}
