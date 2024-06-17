package main

import (
	"encoding/json"
<<<<<<< HEAD
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
=======
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
>>>>>>> f6212a6 (Added Go mod support; basic visualization)
)

type SyftJSON struct {
	Metadata struct {
		Distro string `json:"syft:distro"`
	} `json:"metadata"`
	Manifests map[string]struct {
		Resolved map[string]struct {
			PackageURL string `json:"package_url"`
		} `json:"resolved"`
	} `json:"manifests"`
}

<<<<<<< HEAD
type DiveJSON struct {
	Layers []struct {
		Index     int    `json:"index"`
		ID        string `json:"id"`
		DigestID  string `json:"digestId"`
		SizeBytes int64  `json:"sizeBytes"` // to represent GB sizes
		Command   string `json:"command"`
	} `json:"layer"`
}

func main() {
	// if len(os.Args) < 2 {
	// 	fmt.Println("Usage: syft-analyzer <syft_output_file1.json> ... <syft_output_fileN.json>")
	// 	os.Exit(1)
	// }

	diveFiles := flag.String("dive-files", "", "list of dive files separated by whitespace")
	syftFiles := flag.String("syft-files", "", "list of syft-files separated by whitespace")

	// Parse the flags
	flag.Parse()

	syftFilesSlc := strings.Split(*syftFiles, " ")
	diveFilesSlc := strings.Split(*diveFiles, " ")

	distinctOS := make(map[string]map[string]struct{ size string })
	distinctPackages := make(map[string]int)
	distinctLanguages := make(map[string]int)
	distinctRuntimes := make(map[string]int)

	// parse the syft files and gather usefull information
	for idx := range syftFilesSlc {
		syftData, err := os.ReadFile(syftFilesSlc[idx])
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		diveData, err := os.ReadFile(diveFilesSlc[idx])
=======
type Stats struct {
	BaseOS   map[string]int
	Packages map[string]int
	Runtimes map[string]int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: syft-analyzer <syft_output_file1.json> ... <syft_output_fileN.json>")
		os.Exit(1)
	}

	// Aggregate information from Syft
	stats := Stats{
		BaseOS:   make(map[string]int),
		Packages: make(map[string]int),
		Runtimes: make(map[string]int),
	}

	for _, file := range os.Args[1:] {
		data, err := os.ReadFile(file)
>>>>>>> f6212a6 (Added Go mod support; basic visualization)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		var syftOutput SyftJSON
<<<<<<< HEAD
		err = json.Unmarshal(syftData, &syftOutput)
		if err != nil {
			fmt.Printf("Error parsing syft JSON: %v\n", err)
			os.Exit(1)
		}

		var diveOutput DiveJSON
		err = json.Unmarshal(diveData, &diveOutput)
		if err != nil {
			fmt.Printf("Error parsing dive JSON: %v\n", err)
			os.Exit(1)
		}

		// Aggregate base OS statistics
		osName, version := extractOSNameAndVersion(syftOutput.Metadata.Distro)
		if distinctOS[osName] == nil {
			distinctOS[osName] = make(map[string]struct{ size string })
		}
		distinctOS[osName][version] = struct{ size string }{convertBytesToHumanReadableString(diveOutput.Layers[0].SizeBytes)}

		// Aggregate package statistics
		for _, manifest := range syftOutput.Manifests {
			for _, pkg := range manifest.Resolved {
				distinctPackages[pkg.PackageURL]++
				detectLanguageAndRuntime(pkg.PackageURL, distinctLanguages, distinctRuntimes)
			}
		}
	}

	// Print statistics
	fmt.Println("Base OS Statistics:")
	for os, versionsSet := range distinctOS {
		versions := make([]string, 0, len(versionsSet))
		osVersionSizes := make([]string, 0, len(versionsSet))
		for version := range versionsSet {
			versions = append(versions, version)
			osVersionSizes = append(osVersionSizes, versionsSet[version].size)
		}
		fmt.Printf("OS: %s\n", os)
		fmt.Printf("%d (versions: %s)\n", len(versionsSet), strings.Join(versions, ", "))
		fmt.Printf("%d (sizes: %s)\n", len(versionsSet), strings.Join(osVersionSizes, ", "))
	}

	fmt.Println("\nPackage Statistics:")
	for pkgURL, count := range distinctPackages {
		pkgName := extractPackageName(pkgURL)
		fmt.Printf("%s: %d\n", pkgName, count)
	}

	fmt.Println("\nRuntime Statistics:")
	for runtime, count := range distinctRuntimes {
		fmt.Printf("%s: %d\n", runtime, count)
	}

	// If we reach this point, all checks passed
	fmt.Println("All checks passed.")
}

// extractOSNameAndVersion parses the distro string to extract the OS name and version
func extractOSNameAndVersion(distro string) (osName, version string) {
	// Regular expression to match the OS name and version
	re := regexp.MustCompile(`pkg:generic/([^@]+)@([^?]+)`)
	matches := re.FindStringSubmatch(distro)
	if len(matches) > 2 {
		osName = matches[1]
		version = matches[2]
	}
	return
}

// detectLanguageAndRuntime updates the language and runtime statistics based on the package URL
func detectLanguageAndRuntime(pkgURL string, languages map[string]int, runtimes map[string]int) {
	if strings.Contains(pkgURL, "pkg:generic/java/") || strings.Contains(pkgURL, "pkg:maven/") {
		languages["Java"]++
=======
		err = json.Unmarshal(data, &syftOutput)
		if err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			os.Exit(1)
		}

		osNameWithVersion := extractOSNameWithVersion(syftOutput.Metadata.Distro)
		stats.BaseOS[osNameWithVersion]++

		for _, manifest := range syftOutput.Manifests {
			for _, pkg := range manifest.Resolved {
				stats.Packages[extractPackageName(pkg.PackageURL)]++
				detectRuntime(pkg.PackageURL, stats.Runtimes)
			}
		}
	}
	printMapNicely(stats.Packages)

	// Generate statistics
	duplicatePackages := getDuplicates(stats.Packages)
	printMapNicely(duplicatePackages)
	if err := generateStats(duplicatePackages, "Package Statistics", "stats_packages.png", 10); err != nil {
		fmt.Printf("Error generating bar chart for packages: %v\n", err)
		os.Exit(1)
	}

	duplicateRuntimes := getDuplicates(stats.Runtimes)
	printMapNicely(duplicateRuntimes)
	if err := generateStats(duplicateRuntimes, "Runtime Statistics", "stats_runtime.png", 10); err != nil {
		fmt.Printf("Error generating bar chart for runtimes: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Finished successfully.")
}

func printMapNicely(entries map[string]int) {
	fmt.Println("Entries print:")
	fmt.Println(strings.Repeat("-", 40))

	for key, count := range entries {
		fmt.Printf("%-30s\t%d\n", key, count)
	}
}

func getDuplicates(entries map[string]int) map[string]int {
	duplicates := make(map[string]int)
	for key, count := range entries {
		if count >= 2 {
			duplicates[key] = count
		}
	}
	return duplicates
}

func extractOSNameWithVersion(distro string) (osNameWithVersion string) {
	re := regexp.MustCompile(`pkg:generic/([^@]+)@([^?]+)`)
	matches := re.FindStringSubmatch(distro)
	if len(matches) > 2 {
		osName := matches[1]
		version := matches[2]
		osNameWithVersion = osName + ":" + version
	}
	return osNameWithVersion
}

func detectRuntime(pkgURL string, runtimes map[string]int) {
	if strings.Contains(pkgURL, "pkg:generic/java/") || strings.Contains(pkgURL, "pkg:maven/") {
>>>>>>> f6212a6 (Added Go mod support; basic visualization)
		if strings.Contains(pkgURL, "jre") || strings.Contains(pkgURL, "jdk") {
			runtimes["Java"]++
		}
	} else if strings.Contains(pkgURL, "pkg:python/") {
<<<<<<< HEAD
		languages["Python"]++
=======
>>>>>>> f6212a6 (Added Go mod support; basic visualization)
		if strings.Contains(pkgURL, "cpython") {
			runtimes["Python"]++
		}
	}
<<<<<<< HEAD
	// TODO: Add more language and runtime detection rules as needed
}

// extractPackageName cleans up the package URL to a more readable format
func extractPackageName(pkgURL string) string {
	// Remove the query parameters
	cleanURL := strings.Split(pkgURL, "?")[0]
	// Remove the "pkg:" prefix
	cleanURL = strings.TrimPrefix(cleanURL, "pkg:")
	// Replace URL encoding with actual characters
	cleanURL = strings.ReplaceAll(cleanURL, "%2B", "+")
	cleanURL = strings.ReplaceAll(cleanURL, "%2F", "/")
	return cleanURL
}

func convertBytesToHumanReadableString(sizeBytes int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}

	if sizeBytes == 0 {
		return "0B"
	}
	size := float64(sizeBytes)
	i := 0
	for size >= 1024 && i < len(units)-1 {
		size /= 1024 // Divide size by 1024 to convert to the next higher unit
		i++
	}

	return fmt.Sprintf("%.2f%s", size, units[i])
=======
	// TODO: Add more runtime detection rules as needed, e.g. Golang, Node.js, C/C++, C# ...
}

func extractPackageName(pkgURL string) string {
	parts := strings.Split(pkgURL, "/")
	nameWithVersionRaw := parts[len(parts)-1]
	nameWithVersionFullUrlEncoded := strings.Split(nameWithVersionRaw, "?")[0]

	nameWithVersionUrlDecoded, err := url.QueryUnescape(nameWithVersionFullUrlEncoded)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to decode package URL: %s\n", pkgURL)
		return ""
	}

	return nameWithVersionUrlDecoded
}

func generateStats(stats map[string]int, title string, filename string, topN int) error {
	p := plot.New()
	p.Title.Text = title

	// Create a slice of struct to sort map entries by value
	type kv struct {
		Key   string
		Value int
	}
	var ss []kv
	for k, v := range stats {
		ss = append(ss, kv{k, v})
	}

	// Sort the slice by value, and then by key if values are equal
	sort.Slice(ss, func(i, j int) bool {
		if ss[i].Value == ss[j].Value {
			return ss[i].Key < ss[j].Key
		}
		return ss[i].Value > ss[j].Value
	})

	// If topN is greater than the number of stats, adjust topN to the number of stats
	if topN > len(ss) {
		topN = len(ss)
	}

	// Take only the top N entries for the chart
	ss = ss[:topN]

	// Create the bars and labels for the top N entries
	bars := make(plotter.Values, topN)
	labels := make([]string, topN)
	for i := 0; i < topN; i++ {
		bars[i] = float64(ss[i].Value)
		labels[i] = ss[i].Key
	}

	bar, err := plotter.NewBarChart(bars, vg.Points(20))
	if err != nil {
		return err
	}
	bar.Horizontal = true // Make the bar chart horizontal

	p.Add(bar)
	p.NominalY(labels...)

	// Save the plot to a PNG file.
	if err := p.Save(6*vg.Inch, vg.Length(float64(topN)*0.5)*vg.Inch, filename); err != nil {
		return err
	}

	return nil
>>>>>>> f6212a6 (Added Go mod support; basic visualization)
}
