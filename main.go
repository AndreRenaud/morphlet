package main

import (
	"flag"
	"fmt"
	"image"
	"log"

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
	/*
		img2Triangulate, err := delaunay.Triangulate(img2Points)
		if err != nil {
			log.Fatalf("Unable to triangulate image 1: %v", err)
		}

		log.Printf("img2 triangulation: %v", img2Triangulate.Triangles)
	*/

	for count := 0; count < *frameCount; count++ {
		log.Printf("Doing count %d/%d", count, *frameCount)
		alpha := float64(count) / float64(*frameCount-1) // Range from 0.0 - 1.0

		dst := image.NewNRGBA(img1.Bounds())

		for i := 0; i < len(img1Triangulate.Triangles); i += 3 {
			s1 := img1Points[img1Triangulate.Triangles[i]]
			s2 := img1Points[img1Triangulate.Triangles[i+1]]
			s3 := img1Points[img1Triangulate.Triangles[i+2]]
			//log.Printf("Triangle %d: (%.2f, %.2f), (%.2f, %.2f), (%.2f, %.2f)", i/3, p1.X, p1.Y, p2.X, p2.Y, p3.X, p3.Y)

			//drawTriangle(img1, image.Point{X: int(s1.X), Y: int(s1.Y)}, image.Point{X: int(s2.X), Y: int(s2.Y)}, image.Point{X: int(s3.X), Y: int(s3.Y)}, color.RGBA{255, 0, 0, 255})

			p1 := img2Points[img1Triangulate.Triangles[i]]
			p2 := img2Points[img1Triangulate.Triangles[i+1]]
			p3 := img2Points[img1Triangulate.Triangles[i+2]]
			//drawTriangle(img2, image.Point{X: int(p1.X), Y: int(p1.Y)}, image.Point{X: int(p2.X), Y: int(p2.Y)}, image.Point{X: int(p3.X), Y: int(p3.Y)}, color.RGBA{0, 0, 255, 255})

			WarpTriangle(img2, dst, [3]delaunay.Point{p1, p2, p3}, [3]delaunay.Point{s1, s2, s3})
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

		/*
			for y := 0; y < combined.Bounds().Dy(); y++ {
				for x := 0; x < combined.Bounds().Dx(); x++ {
					c1 := img1.At(x, y).(color.NRGBA)
					c2 := dst.At(x, y).(color.NRGBA)
					combined.Set(x, y, color.NRGBA{
						R: uint8(float64(c1.R)*(1-alpha) + float64(c2.R)*alpha),
						G: uint8(float64(c1.G)*(1-alpha) + float64(c2.G)*alpha),
						B: uint8(float64(c1.B)*(1-alpha) + float64(c2.B)*alpha),
						A: 255,
					})
				}
			}
		*/

		saveImage(combined, fmt.Sprintf("warped-%05d.png", count))
	}

	/*
		for i := 0; i < len(img2Triangulate.Triangles); i += 3 {
			p1 := img2Points[img2Triangulate.Triangles[i]]
			p2 := img2Points[img2Triangulate.Triangles[i+1]]
			p3 := img2Points[img2Triangulate.Triangles[i+2]]
			log.Printf("Triangle %d: (%.2f, %.2f), (%.2f, %.2f), (%.2f, %.2f)", i/3, p1.X, p1.Y, p2.X, p2.Y, p3.X, p3.Y)

			drawTriangle(img2, image.Point{X: int(p1.X), Y: int(p1.Y)}, image.Point{X: int(p2.X), Y: int(p2.Y)}, image.Point{X: int(p3.X), Y: int(p3.Y)}, color.NRGBA{0, 0, 255, 255})
		}
	*/
}
