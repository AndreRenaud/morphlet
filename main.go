package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"sync"

	"github.com/AndreRenaud/morphlet/warp"
	"github.com/fogleman/delaunay"

	"image/draw"
	_ "image/jpeg"
)

func main() {
	frameCount := flag.Int("frames", 21, "Number of frames to generate")
	flag.Parse()

	image1Name := "Newall Timelapse Photos/20180309_183059.jpg"
	image2Name := "Newall Timelapse Photos/20190123_121335.jpg"

	points := [][2]delaunay.Point{
		{{1205, 974}, {1026, 1066}},  // Top left
		{{2573, 793}, {2430, 919}},   // top right
		{{1167, 1916}, {960, 2067}},  // bottom left
		{{2580, 1966}, {2393, 2117}}, // bottom right
	}

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

	if img1.Bounds().Dx() != img2.Bounds().Dx() || img1.Bounds().Dy() != img2.Bounds().Dy() {
		log.Fatalf("images must be of the same size: img1 is %v, img2 is %v", img1.Bounds(), img2.Bounds())
	}

	img1Points := []delaunay.Point{{0, 0}, {float64(img1.Bounds().Dx() - 1), 0},
		{0, float64(img1.Bounds().Dy() - 1)}, {float64(img1.Bounds().Dx() - 1), float64(img1.Bounds().Dy() - 1)}}

	for _, p := range points {
		img1Points = append(img1Points, p[0])
	}
	img2Points := []delaunay.Point{{0, 0}, {float64(img2.Bounds().Dx() - 1), 0},
		{0, float64(img2.Bounds().Dy() - 1)}, {float64(img2.Bounds().Dx() - 1), float64(img2.Bounds().Dy() - 1)}}
	for _, p := range points {
		img2Points = append(img2Points, p[1])
	}

	img1Triangulate, err := delaunay.Triangulate(img1Points)
	if err != nil {
		log.Fatalf("Unable to triangulate image 1: %v", err)
	}
	log.Printf("img1 triangulation: %v", img1Triangulate.Triangles)
	parallel := sync.WaitGroup{}

	for count := 0; count < *frameCount; count++ {
		filename := fmt.Sprintf("warped-%05d.png", count)

		parallel.Go(func() {
			alpha := float64(count) / float64(*frameCount-1) // Range from 0.0 - 1.0

			sourcePoints := make([]delaunay.Point, len(img1Triangulate.Triangles))
			destPoints := make([]delaunay.Point, len(img1Triangulate.Triangles))

			for i := range img1Triangulate.Triangles {
				sourcePoints[i] = img1Points[img1Triangulate.Triangles[i]]
				destPoints[i] = img2Points[img1Triangulate.Triangles[i]]
			}

			dst, err := warp.WarpImage(img2, sourcePoints, destPoints)
			if err != nil {
				log.Fatalf("Cannot warp image: %s", err)
			}

			// Blend img1 with dst at alpha ratio
			combined := image.NewNRGBA(dst.Bounds())
			draw.Draw(combined, combined.Bounds(), img1, image.Point{0, 0}, draw.Src)
			// Set the alpha on dst to alpha
			alphaInt := uint8(255 * alpha)
			for y := 0; y < dst.Bounds().Dy(); y++ {
				for x := 0; x < dst.Bounds().Dx(); x++ {
					dst.Pix[y*dst.Stride+x*4+3] = alphaInt
				}
			}
			draw.Draw(combined, combined.Bounds(), dst, image.Point{0, 0}, draw.Over)
			saveImage(combined, filename)
			log.Printf("Finished %s", filename)
		})
	}
	parallel.Wait()
}
