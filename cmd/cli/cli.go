package main

import (
	"flag"
	"log"

	"github.com/AndreRenaud/morphlet/warp"

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
	if err := job.Run("warped", *frameCount); err != nil {
		log.Fatalf("failed to run warp job: %v", err)
	}
}
