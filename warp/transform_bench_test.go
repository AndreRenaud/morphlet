package warp

import (
	"image"
	"image/color"
	"testing"

	"github.com/fogleman/delaunay"
)

// createTestImage creates a test image with a gradient pattern for benchmarking
func createTestImage(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create a gradient pattern that's interesting to sample
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(((x + y) * 255) / (width + height))
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	return img
}

// BenchmarkWarpTriangle benchmarks the basic triangle warping function
func BenchmarkWarpTriangle(b *testing.B) {
	// Create test images
	srcImg := createTestImage(512, 512)
	dstImg := image.NewNRGBA(image.Rect(0, 0, 512, 512))

	// Define source and destination triangles
	source := [3]delaunay.Point{
		{50, 50},
		{200, 60},
		{100, 180},
	}
	dest := [3]delaunay.Point{
		{60, 40},
		{220, 80},
		{90, 200},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := WarpTriangle(srcImg, dstImg, source, dest)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWarpTriangleLarge benchmarks with a larger triangle
func BenchmarkWarpTriangleLarge(b *testing.B) {
	srcImg := createTestImage(1024, 1024)
	dstImg := image.NewNRGBA(image.Rect(0, 0, 1024, 1024))

	// Larger triangle covering more of the image
	source := [3]delaunay.Point{
		{100, 100},
		{800, 150},
		{400, 700},
	}
	dest := [3]delaunay.Point{
		{120, 80},
		{850, 200},
		{350, 750},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := WarpTriangle(srcImg, dstImg, source, dest)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWarpTriangleSmall benchmarks with a very small triangle
func BenchmarkWarpTriangleSmall(b *testing.B) {
	srcImg := createTestImage(256, 256)
	dstImg := image.NewNRGBA(image.Rect(0, 0, 256, 256))

	// Small triangle
	source := [3]delaunay.Point{
		{10, 10},
		{30, 15},
		{20, 35},
	}
	dest := [3]delaunay.Point{
		{15, 8},
		{35, 18},
		{18, 40},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := WarpTriangle(srcImg, dstImg, source, dest)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSampleBilinear benchmarks the bilinear sampling function separately
func BenchmarkSampleBilinear(b *testing.B) {
	srcImg := createTestImage(512, 512)

	// Test coordinates for sampling
	coords := []struct{ x, y float64 }{
		{100.5, 100.5},
		{255.7, 255.3},
		{400.2, 300.8},
		{50.1, 450.9},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, coord := range coords {
			_ = sampleBilinear(srcImg, coord.x, coord.y)
		}
	}
}

// BenchmarkWarpImage benchmarks the full image warping with multiple triangles
func BenchmarkWarpImage(b *testing.B) {
	srcImg := createTestImage(512, 512)

	// Create a simple set of triangles (2 triangles forming a quad)
	sourcePoints := []delaunay.Point{
		// Triangle 1
		{50, 50},
		{250, 50},
		{50, 250},
		// Triangle 2
		{250, 50},
		{250, 250},
		{50, 250},
	}

	destPoints := []delaunay.Point{
		// Triangle 1 (slightly warped)
		{60, 40},
		{260, 60},
		{40, 260},
		// Triangle 2 (slightly warped)
		{260, 60},
		{240, 270},
		{40, 260},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := WarpImage(srcImg, sourcePoints, destPoints)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMatrixOperations benchmarks the matrix math operations
func BenchmarkMatrixOperations(b *testing.B) {
	// Test matrices
	M := [2][2]float64{
		{1.5, 0.3},
		{-0.2, 2.1},
	}
	v := delaunay.Point{100.5, 200.7}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test matrix inverse
		_, ok := inv2x2(M)
		if !ok {
			b.Fatal("Matrix inversion failed")
		}

		// Test matrix-vector multiplication
		_ = mul2(M, v)

		// Test vector operations
		_ = add(v, delaunay.Point{10, 20})
		_ = sub(v, delaunay.Point{5, 15})
	}
}
