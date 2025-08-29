package warp

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/fogleman/delaunay"
)

func sub(a, b delaunay.Point) delaunay.Point { return delaunay.Point{a.X - b.X, a.Y - b.Y} }
func add(a, b delaunay.Point) delaunay.Point { return delaunay.Point{a.X + b.X, a.Y + b.Y} }
func mul2(M [2][2]float64, v delaunay.Point) delaunay.Point {
	return delaunay.Point{
		M[0][0]*v.X + M[0][1]*v.Y,
		M[1][0]*v.X + M[1][1]*v.Y,
	}
}

func inv2x2(M [2][2]float64) (Minv [2][2]float64, ok bool) {
	det := M[0][0]*M[1][1] - M[0][1]*M[1][0]
	if math.Abs(det) < 1e-12 {
		return Minv, false
	}
	id := 1.0 / det
	Minv[0][0] = M[1][1] * id
	Minv[0][1] = -M[0][1] * id
	Minv[1][0] = -M[1][0] * id
	Minv[1][1] = M[0][0] * id
	return Minv, true
}

// Convert any image.Image to *image.RGBA once for fast random access.
func toRGBA(src image.Image) *image.RGBA {
	rgba := image.NewRGBA(src.Bounds())
	draw.Draw(rgba, rgba.Bounds(), src, src.Bounds().Min, draw.Src)
	return rgba
}

// Bilinear sampling from *image.RGBA at floating point (x,y).
func sampleBilinear(rgba *image.RGBA, x, y float64) color.RGBA {
	b := rgba.Bounds()
	if x < float64(b.Min.X) {
		x = float64(b.Min.X)
	}
	if y < float64(b.Min.Y) {
		y = float64(b.Min.Y)
	}
	if x > float64(b.Max.X-1) {
		x = float64(b.Max.X - 1)
	}
	if y > float64(b.Max.Y-1) {
		y = float64(b.Max.Y - 1)
	}

	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1 := x0 + 1
	y1 := y0 + 1
	if x1 >= b.Max.X {
		x1 = b.Max.X - 1
	}
	if y1 >= b.Max.Y {
		y1 = b.Max.Y - 1
	}

	fx := x - float64(x0)
	fy := y - float64(y0)

	c00 := rgba.RGBAAt(x0, y0)
	c10 := rgba.RGBAAt(x1, y0)
	c01 := rgba.RGBAAt(x0, y1)
	c11 := rgba.RGBAAt(x1, y1)

	lerp := func(a, b float64, t float64) float64 { return a + (b-a)*t }

	r00 := float64(c00.R)
	r10 := float64(c10.R)
	r01 := float64(c01.R)
	r11 := float64(c11.R)
	g00 := float64(c00.G)
	g10 := float64(c10.G)
	g01 := float64(c01.G)
	g11 := float64(c11.G)
	b00 := float64(c00.B)
	b10 := float64(c10.B)
	b01 := float64(c01.B)
	b11 := float64(c11.B)
	a00 := float64(c00.A)
	a10 := float64(c10.A)
	a01 := float64(c01.A)
	a11 := float64(c11.A)

	r0 := lerp(r00, r10, fx)
	r1 := lerp(r01, r11, fx)
	g0 := lerp(g00, g10, fx)
	g1 := lerp(g01, g11, fx)
	b0 := lerp(b00, b10, fx)
	b1 := lerp(b01, b11, fx)
	a0 := lerp(a00, a10, fx)
	a1 := lerp(a01, a11, fx)

	R := uint8(math.Round(lerp(r0, r1, fy)))
	G := uint8(math.Round(lerp(g0, g1, fy)))
	B := uint8(math.Round(lerp(b0, b1, fy)))
	A := uint8(math.Round(lerp(a0, a1, fy)))

	return color.RGBA{R, G, B, A}
}

// WarpTriangle maps src triangle S -> dst triangle D by filling into dst image.
// srcImg may be any image.Image; dst must be draw.Image (e.g. *image.RGBA).
func WarpTriangle(srcImg image.Image, dst draw.Image, S [3]delaunay.Point, D [3]delaunay.Point) error {
	src := toRGBA(srcImg) // one-time conversion for speed

	// Build edge matrices: P = [S1-S0 | S2-S0], Q = [D1-D0 | D2-D0]
	P := [2][2]float64{
		{S[1].X - S[0].X, S[2].X - S[0].X},
		{S[1].Y - S[0].Y, S[2].Y - S[0].Y},
	}
	Q := [2][2]float64{
		{D[1].X - D[0].X, D[2].X - D[0].X},
		{D[1].Y - D[0].Y, D[2].Y - D[0].Y},
	}

	Qinv, ok := inv2x2(Q)
	if !ok {
		return nil // Degenerate destination triangle; nothing to do.
	}

	// Precompute bounding box of the destination triangle to scan.
	minX := int(math.Floor(math.Min(D[0].X, math.Min(D[1].X, D[2].X))))
	maxX := int(math.Ceil(math.Max(D[0].X, math.Max(D[1].X, D[2].X))))
	minY := int(math.Floor(math.Min(D[0].Y, math.Min(D[1].Y, D[2].Y))))
	maxY := int(math.Ceil(math.Max(D[0].Y, math.Max(D[1].Y, D[2].Y))))

	dstBounds := dst.Bounds()
	if minX < dstBounds.Min.X {
		minX = dstBounds.Min.X
	}
	if minY < dstBounds.Min.Y {
		minY = dstBounds.Min.Y
	}
	if maxX > dstBounds.Max.X {
		maxX = dstBounds.Max.X
	}
	if maxY > dstBounds.Max.Y {
		maxY = dstBounds.Max.Y
	}

	// Scanline over the destination bounding box.
	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			// Center of pixel for nicer results
			p := delaunay.Point{float64(x) + 0.5, float64(y) + 0.5}

			// Compute barycentric-like coords (u,v) solving Q * [u v]^T = (p - D0)
			rel := sub(p, D[0])
			uv := mul2(Qinv, rel)
			u, v := uv.X, uv.Y
			w := 1.0 - u - v

			// Inside test for triangle (allow tiny epsilon on edges)
			if u >= -1e-6 && v >= -1e-6 && w >= -1e-6 {
				// Map to source: s = S0 + P * [u v]
				srcRel := delaunay.Point{
					P[0][0]*u + P[0][1]*v,
					P[1][0]*u + P[1][1]*v,
				}
				s := add(S[0], srcRel)

				c := sampleBilinear(src, s.X, s.Y)
				dst.Set(x, y, c)
			}
		}
	}
	return nil
}

func WarpImage(srcImg image.Image, sourcePoints, destPoints []delaunay.Point) (*image.NRGBA, error) {
	if len(sourcePoints) != len(destPoints) {
		return nil, fmt.Errorf("source and destination point lists must have the same length")
	}
	if len(sourcePoints)%3 != 0 {
		return nil, fmt.Errorf("point lists length must be a multiple of 3")
	}
	dstImg := image.NewNRGBA(srcImg.Bounds())

	for i := 0; i < len(sourcePoints); i += 3 {
		//for i := 0; i < len(img1Triangulate.Triangles); i += 3 {
		//s1 := img1Points[img1Triangulate.Triangles[i]]
		//s2 := img1Points[img1Triangulate.Triangles[i+1]]
		//s3 := img1Points[img1Triangulate.Triangles[i+2]]
		//log.Printf("Triangle %d: (%.2f, %.2f), (%.2f, %.2f), (%.2f, %.2f)", i/3, p1.X, p1.Y, p2.X, p2.Y, p3.X, p3.Y)

		//drawTriangle(img1, image.Point{X: int(s1.X), Y: int(s1.Y)}, image.Point{X: int(s2.X), Y: int(s2.Y)}, image.Point{X: int(s3.X), Y: int(s3.Y)}, color.RGBA{255, 0, 0, 255})

		//p1 := img2Points[img1Triangulate.Triangles[i]]
		//p2 := img2Points[img1Triangulate.Triangles[i+1]]
		//p3 := img2Points[img1Triangulate.Triangles[i+2]]
		//drawTriangle(img2, image.Point{X: int(p1.X), Y: int(p1.Y)}, image.Point{X: int(p2.X), Y: int(p2.Y)}, image.Point{X: int(p3.X), Y: int(p3.Y)}, color.RGBA{0, 0, 255, 255})

		source := [3]delaunay.Point{sourcePoints[i], sourcePoints[i+1], sourcePoints[i+2]}
		dest := [3]delaunay.Point{destPoints[i], destPoints[i+1], destPoints[i+2]}

		WarpTriangle(srcImg, dstImg, dest, source)
	}
	return dstImg, nil
}
