package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"ova-size-optimizer/logic/aggregate"
	"ova-size-optimizer/logic/load"
	"ova-size-optimizer/logic/visualize"
)

func main() {
	diveFiles := flag.String("dive-files", "", "list of dive files separated by whitespace")
	syftFiles := flag.String("syft-files", "", "list of syft-files separated by whitespace")

	flag.Parse()

	syftFilesSlc := strings.Split(*syftFiles, " ")
	diveFilesSlc := strings.Split(*diveFiles, " ")

	stats := aggregate.NewStats()

	for idx := range len(syftFilesSlc) {
		syftOutput, err := load.LoadSyftFile(syftFilesSlc[idx])
		if err != nil {
			fmt.Printf("Error processing Syft file: %v\n", err)
			os.Exit(1)
		}

		diveOutput, err := load.LoadDiveFile(diveFilesSlc[idx])
		if err != nil {
			fmt.Printf("Error processing Dive file: %v\n", err)
			os.Exit(1)
		}

		aggregate.ProcessData(stats, syftOutput, diveOutput)
	}

	// aggregate.debugMapPrint(stats.BaseOS)
	// aggregate.debugMapPrint(stats.Packages)
	// aggregate.debugMapPrint(stats.Runtimes)

	duplicateBaseOS := aggregate.GetOnlyDuplicates(stats.BaseOS)
	duplicatePackages := aggregate.GetOnlyDuplicates(stats.Packages)
	duplicateRuntimes := aggregate.GetOnlyDuplicates(stats.Runtimes)

	if err := visualize.PlotStats(duplicateBaseOS, "BaseOS Statistics", "stats_base_os.png", 10); err != nil {
		fmt.Printf("Error generating bar chart for BaseOS: %v\n", err)
		os.Exit(1)
	}

	if err := visualize.PlotStats(duplicatePackages, "Package Statistics", "stats_packages.png", 10); err != nil {
		fmt.Printf("Error generating bar chart for packages: %v\n", err)
		os.Exit(1)
	}

	if err := visualize.PlotStats(duplicateRuntimes, "Runtime Statistics", "stats_runtimes.png", 10); err != nil {
		fmt.Printf("Error generating bar chart for runtimes: %v\n", err)
		os.Exit(1)
	}

	// aggregate.DebugMapPrint(duplicateBaseOS)
	// aggregate.DebugMapPrint(duplicatePackages)
	// aggregate.DebugMapPrint(duplicateRuntimes)

	fmt.Println("Finished successfully.")
}
