package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
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
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		var syftOutput SyftJSON
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
		if strings.Contains(pkgURL, "jre") || strings.Contains(pkgURL, "jdk") {
			runtimes["Java"]++
		}
	} else if strings.Contains(pkgURL, "pkg:python/") {
		languages["Python"]++
		if strings.Contains(pkgURL, "cpython") {
			runtimes["Python"]++
		}
	}
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
}
