package warp

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/fogleman/delaunay"
)

type WarpJob struct {
	Images      []*image.NRGBA
	ImagePoints [][]delaunay.Point
	ThreadCount int // Number of concurrent threads to use. If set to 0, uses auto detected CPU count
}

type WarpJobSaveFormat struct {
	Images      []string  `json:"images"`
	ImagePoints [][][]int `json:"image_points"`
}

func (w *WarpJob) Run(filePrefix string, frameCount int) error {
	if len(w.Images) < 2 {
		return fmt.Errorf("need at least two images to warp")
	}
	if len(w.ImagePoints) != len(w.Images) {
		return fmt.Errorf("need the same number of image points as images (have %d images, %d image points)", len(w.Images), len(w.ImagePoints))
	}
	for i := 1; i < len(w.Images); i++ {
		if w.Images[i].Bounds().Dx() != w.Images[0].Bounds().Dx() || w.Images[i].Bounds().Dy() != w.Images[0].Bounds().Dy() {
			return fmt.Errorf("all images must be of the same size: image 0 is %v, image %d is %v", w.Images[0].Bounds(), i, w.Images[i].Bounds())
		}
	}
	for i := 1; i < len(w.ImagePoints); i++ {
		if len(w.ImagePoints[i]) != len(w.ImagePoints[0]) {
			return fmt.Errorf("need the same number of points for all images. image0 has %d, image %d has %d", len(w.ImagePoints[0]), i, len(w.ImagePoints[i]))
		}
	}
	bounds := w.Images[0].Bounds()
	img0Points := []delaunay.Point{{0, 0}, {float64(bounds.Dx() - 1), 0},
		{0, float64(bounds.Dy() - 1)}, {float64(bounds.Dx() - 1), float64(bounds.Dy() - 1)}}
	img0Points = append(img0Points, w.ImagePoints[0]...)

	triangulate, err := delaunay.Triangulate(img0Points)
	if err != nil {
		return fmt.Errorf("unable to triangulate image 1: %v", err)
	}
	log.Printf("triangulation: %v", triangulate.Triangles)

	fileCount := 0
	prevImage := w.Images[0]
	if w.ThreadCount <= 0 {
		w.ThreadCount = runtime.NumCPU()
	}
	jobCount := make(chan struct{}, w.ThreadCount)
	for imageIdx := 1; imageIdx < len(w.Images); imageIdx++ {

		parallel := sync.WaitGroup{}
		var prevImageCandidate *image.NRGBA
		for count := 0; count < frameCount; count++ {
			filename := fmt.Sprintf("%s-%05d.png", filePrefix, fileCount)
			fileCount++
			jobCount <- struct{}{}
			parallel.Go(func() {

				alpha := float64(count) / float64(frameCount-1) // Range from 0.0 - 1.0

				sourcePoints := make([]delaunay.Point, len(triangulate.Triangles))
				destPoints := make([]delaunay.Point, len(triangulate.Triangles))

				imgPoints := []delaunay.Point{{0, 0}, {float64(bounds.Dx() - 1), 0},
					{0, float64(bounds.Dy() - 1)}, {float64(bounds.Dx() - 1), float64(bounds.Dy() - 1)}}
				imgPoints = append(imgPoints, w.ImagePoints[imageIdx]...)

				for i := range triangulate.Triangles {
					sourcePoints[i] = img0Points[triangulate.Triangles[i]]
					destPoints[i] = imgPoints[triangulate.Triangles[i]]
				}

				dst, err := WarpImage(w.Images[imageIdx], sourcePoints, destPoints)
				if err != nil {
					log.Fatalf("Cannot warp image: %s", err)
				}

				// Blend img1 with dst at alpha ratio
				combined := image.NewNRGBA(dst.Bounds())
				draw.Draw(combined, combined.Bounds(), prevImage, image.Point{0, 0}, draw.Src)
				// Set the alpha on dst to alpha
				alphaInt := uint8(255 * alpha)
				for y := 0; y < dst.Bounds().Dy(); y++ {
					for x := 0; x < dst.Bounds().Dx(); x++ {
						dst.Pix[y*dst.Stride+x*4+3] = alphaInt
					}
				}
				draw.Draw(combined, combined.Bounds(), dst, image.Point{0, 0}, draw.Over)
				SaveImage(combined, filename)
				log.Printf("Finished %s", filename)

				if count == frameCount-1 {
					prevImageCandidate = dst
				}
				<-jobCount
			})
		}
		parallel.Wait()
		prevImage = prevImageCandidate
	}
	return nil
}

func LoadWarpJson(filename string) (*WarpJobSaveFormat, error) {
	var saved WarpJobSaveFormat
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonData, &saved); err != nil {
		return nil, err
	}
	log.Printf("Loaded warp job: %+v", saved)
	return &saved, nil
}

func SaveWarpJson(job *WarpJobSaveFormat, filename string) error {
	jsonData, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal job data: %w", err)
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("Saved warp job to: %s", filename)
	return nil
}

func NewJobFromFile(filename string) (*WarpJob, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewJobFromJson(data)
}

func NewJobFromJson(jsonData []byte) (*WarpJob, error) {
	var saved WarpJobSaveFormat
	if err := json.Unmarshal(jsonData, &saved); err != nil {
		return nil, err
	}

	var job WarpJob
	for _, imagePoints := range saved.ImagePoints {
		var points []delaunay.Point
		for _, point := range imagePoints {
			if len(point) != 2 {
				return nil, fmt.Errorf("invalid point format: %v", point)
			}
			points = append(points, delaunay.Point{X: float64(point[0]), Y: float64(point[1])})
		}
		job.ImagePoints = append(job.ImagePoints, points)
	}
	for _, imageName := range saved.Images {
		img, err := LoadImage(imageName)

		if err != nil {
			return nil, err
		}
		job.Images = append(job.Images, img)
	}
	return &job, nil
}
