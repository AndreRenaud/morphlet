package main

import (
	"image"
	"log"

	"github.com/fogleman/delaunay"

	_ "image/jpeg"
)

func main() {
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

	dst := image.NewRGBA(img1.Bounds())

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
	saveImage(img1, "img1_triangles.png")
	saveImage(img2, "img2_triangles.png")
	saveImage(dst, "warped.png")

	/*
		for i := 0; i < len(img2Triangulate.Triangles); i += 3 {
			p1 := img2Points[img2Triangulate.Triangles[i]]
			p2 := img2Points[img2Triangulate.Triangles[i+1]]
			p3 := img2Points[img2Triangulate.Triangles[i+2]]
			log.Printf("Triangle %d: (%.2f, %.2f), (%.2f, %.2f), (%.2f, %.2f)", i/3, p1.X, p1.Y, p2.X, p2.Y, p3.X, p3.Y)

			drawTriangle(img2, image.Point{X: int(p1.X), Y: int(p1.Y)}, image.Point{X: int(p2.X), Y: int(p2.Y)}, image.Point{X: int(p3.X), Y: int(p3.Y)}, color.RGBA{0, 0, 255, 255})
		}
	*/
}
