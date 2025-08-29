package main

import (
	"flag"
	"log"
	"sync"

	"github.com/AndreRenaud/morphlet/warp"
	progressbar "github.com/schollz/progressbar/v3"

	_ "image/jpeg"
)

func main() {
	frameCount := flag.Int("frames", 21, "Number of frames to generate")
	jobFile := flag.String("job", "", "Json file containing warp job details (see warp/WarpJsonSaveFormat)")
	flag.Parse()

	job, err := warp.NewJobFromFile(*jobFile)
	if err != nil {
		log.Fatalf("Cannot load %s: %s", *jobFile, err)
	}
	var mutex sync.Mutex
	var bar *progressbar.ProgressBar
	progressCB := func(pos int, total int) {
		mutex.Lock()
		defer mutex.Unlock()
		if bar == nil {
			bar = progressbar.Default(int64(total))
		}
		bar.Set(pos)
	}
	job.Callback = progressCB

	if err := job.Run("warped", *frameCount); err != nil {
		log.Fatalf("failed to run warp job: %v", err)
	}
}
