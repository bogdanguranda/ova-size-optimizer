package main

import (
	"fmt"
	"os"

	"ova-size-optimizer/logic/analyze"
	"ova-size-optimizer/logic/ociimage"
	"ova-size-optimizer/logic/visualize"
)

func main() {
	ovaImagePath := os.Args[1]
	if ovaImagePath == "" {
		fmt.Println("Please provide ova path as an argument")
		os.Exit(1)
	}

	if err := ociimage.TransformAndCopyOciToDockerImage(ovaImagePath); err != nil {
		fmt.Printf("error processing OCI image: %v\n", err)
		os.Exit(1)
	}

	archivesStats, err := analyze.Analyze(ociimage.IndividualArchivesDir)
	if err != nil {
		fmt.Printf("error anaylzing individual archives: %v\n", err)
		os.Exit(1)
	}

	// TODO: change this to a HTML report
	err = visualize.GenerateReport(archivesStats)
	if err != nil {
		fmt.Printf("error generating report: %v\n", err)
		os.Exit(1)
	}
}
