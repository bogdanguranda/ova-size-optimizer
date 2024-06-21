package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"ova-size-optimizer/logic/aggregate"
	"ova-size-optimizer/logic/load"
	"ova-size-optimizer/logic/visualize"
)

func main() {
	syftGithubJSONFiles := flag.String("syft-github-json-files", "", "list of syft-files separated by whitespace exported in github-json format")
	syftJSONFiles := flag.String("syft-json-files", "", "list of syft-files separated by whitespace exported in json format")

	flag.Parse()

	syftGithubJSONFilesSlc := strings.Split(*syftGithubJSONFiles, " ")
	syftJSONFilesSlc := strings.Split(*syftJSONFiles, " ")

	stats := aggregate.NewStats()

	for idx := range len(syftGithubJSONFilesSlc) {
		fileName := syftJSONFilesSlc[idx]
		stats.Runtimes[fileName] = make(map[string]*aggregate.Info)
		syftGithubJSONOutput, err := load.LoadGithubJsonSyftFile(syftGithubJSONFilesSlc[idx])
		if err != nil {
			fmt.Printf("Error processing Github Json Syft file: %v\n", err)
			os.Exit(1)
		}

		syftJSONOutput, err := load.LoadJsonSyftFile(syftJSONFilesSlc[idx])
		if err != nil {
			fmt.Printf("Error processing Json Syft file: %v\n", err)
			os.Exit(1)
		}

		aggregate.ProcessData(stats, syftGithubJSONOutput, syftJSONOutput, fileName)
	}

	// final parse
	for _, fileName := range syftJSONFilesSlc {
		for k, v := range stats.Runtimes[fileName] {
			sizeInt64, err := strconv.ParseInt(v.Size, 10, 64)
			if err != nil {
				fmt.Printf("Failed to convert string to int64 %v\n", err)
				os.Exit(1)
			}
			v.Size = aggregate.ConvertSizeBytesToHumanReadableString(sizeInt64)
			stats.Runtimes[fileName][k] = v
		}
	}

	aggregate.DebugMapPrint(stats.BaseOS)
	aggregate.DebugMapPrint(stats.Packages)
	aggregate.DebugRuntimeMapPrint(stats.Runtimes)

	duplicateBaseOS := aggregate.GetOnlyDuplicates(stats.BaseOS)
	duplicatePackages := aggregate.GetOnlyDuplicates(stats.Packages)
	duplicateRuntimes := aggregate.GetOnlyDuplicatesRuntimes(stats.Runtimes)

	if err := visualize.PlotStats(duplicateBaseOS, "BaseOS Statistics", "stats-base-os.png", 10); err != nil {
		fmt.Printf("Error generating bar chart for BaseOS: %v\n", err)
		os.Exit(1)
	}

	if err := visualize.PlotStats(duplicatePackages, "Package Statistics", "stats-packages.png", 10); err != nil {
		fmt.Printf("Error generating bar chart for packages: %v\n", err)
		os.Exit(1)
	}

	if err := visualize.PlotStats(duplicateRuntimes, "Runtime Statistics", "stats-runtimes.png", 10); err != nil {
		fmt.Printf("Error generating bar chart for runtimes: %v\n", err)
		os.Exit(1)
	}

	// aggregate.DebugMapPrint(duplicateBaseOS)
	// aggregate.DebugMapPrint(duplicatePackages)
	// aggregate.DebugMapPrint(duplicateRuntimes)

	fmt.Println("Finished successfully.")
}
