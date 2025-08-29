package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"github.com/fogleman/delaunay"

	_ "image/jpeg"
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

	for i := 0; i < len(img1Triangulate.Triangles); i += 3 {
		p1 := img1Points[img1Triangulate.Triangles[i]]
		p2 := img1Points[img1Triangulate.Triangles[i+1]]
		p3 := img1Points[img1Triangulate.Triangles[i+2]]
		//log.Printf("Triangle %d: (%.2f, %.2f), (%.2f, %.2f), (%.2f, %.2f)", i/3, p1.X, p1.Y, p2.X, p2.Y, p3.X, p3.Y)

		drawTriangle(img1, image.Point{X: int(p1.X), Y: int(p1.Y)}, image.Point{X: int(p2.X), Y: int(p2.Y)}, image.Point{X: int(p3.X), Y: int(p3.Y)}, color.RGBA{255, 0, 0, 255})

		p1 = img2Points[img1Triangulate.Triangles[i]]
		p2 = img2Points[img1Triangulate.Triangles[i+1]]
		p3 = img2Points[img1Triangulate.Triangles[i+2]]
		drawTriangle(img2, image.Point{X: int(p1.X), Y: int(p1.Y)}, image.Point{X: int(p2.X), Y: int(p2.Y)}, image.Point{X: int(p3.X), Y: int(p3.Y)}, color.RGBA{0, 0, 255, 255})
	}
	saveImage(img1, "img1_triangles.png")
	saveImage(img2, "img2_triangles.png")

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
