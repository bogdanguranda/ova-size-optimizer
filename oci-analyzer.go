package main

import (
	"encoding/json"
	"net/url"
	"sort"

	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
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

type Info struct {
	Count int
	Size  string
}

type Stats struct {
	BaseOS   map[string]*Info
	Packages map[string]*Info
	Runtimes map[string]*Info
}

func main() {
	diveFiles := flag.String("dive-files", "", "list of dive files separated by whitespace")
	syftFiles := flag.String("syft-files", "", "list of syft-files separated by whitespace")

	flag.Parse()

	syftFilesSlc := strings.Split(*syftFiles, " ")
	diveFilesSlc := strings.Split(*diveFiles, " ")

	stats := Stats{
		BaseOS:   make(map[string]*Info),
		Packages: make(map[string]*Info),
		Runtimes: make(map[string]*Info),
	}

	for idx := range syftFilesSlc {
		syftData, err := os.ReadFile(syftFilesSlc[idx])
		if err != nil {
			fmt.Printf("Error reading syft JSON: %v\n", err)
			os.Exit(1)
		}

		diveData, err := os.ReadFile(diveFilesSlc[idx])
		if err != nil {
			fmt.Printf("Error reading dive JSON: %v\n", err)
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

		osNameWithVersion := extractOSNameWithVersion(syftOutput.Metadata.Distro)
		if stats.BaseOS[osNameWithVersion] == nil {
			stats.BaseOS[osNameWithVersion] = &Info{
				Count: 1,
				Size:  convertBytesToHumanReadableString(diveOutput.Layers[0].SizeBytes),
			}
		} else {
			stats.BaseOS[osNameWithVersion].Count++
		}

		for _, manifest := range syftOutput.Manifests {
			for _, pkg := range manifest.Resolved {
				packageName := extractPackageName(pkg.PackageURL)
				if stats.Packages[packageName] == nil {
					stats.Packages[packageName] = &Info{
						Count: 1,
						Size:  "0", // TODO: implement size determination for packages
					}
				} else {
					stats.Packages[packageName].Count++
				}

				detectRuntime(pkg.PackageURL, stats.Runtimes)
			}
		}
	}

	// debugMapPrint(stats.BaseOS)
	// debugMapPrint(stats.Packages)
	// debugMapPrint(stats.Runtimes)

	// Generate statistics
	duplicateBaseOS := getDuplicates(stats.BaseOS)
	if err := generateStats(duplicateBaseOS, "BaseOS Statistics", "stats_base_os.png", 10); err != nil {
		fmt.Printf("Error generating bar chart for packages: %v\n", err)
		os.Exit(1)
	}

	duplicatePackages := getDuplicates(stats.Packages)
	if err := generateStats(duplicatePackages, "Package Statistics", "stats_packages.png", 10); err != nil {
		fmt.Printf("Error generating bar chart for packages: %v\n", err)
		os.Exit(1)
	}

	duplicateRuntimes := getDuplicates(stats.Runtimes)
	if err := generateStats(duplicateRuntimes, "Runtime Statistics", "stats_runtime.png", 10); err != nil {
		fmt.Printf("Error generating bar chart for runtimes: %v\n", err)
		os.Exit(1)
	}

	debugMapPrint(duplicateBaseOS)
	debugMapPrint(duplicatePackages)
	debugMapPrint(duplicateRuntimes)

	fmt.Println("Finished successfully.")
}

func debugMapPrint(entries map[string]*Info) {
	fmt.Println("Entries print:")
	fmt.Println(strings.Repeat("-", 40))

	for key, info := range entries {
		fmt.Printf("%-30s\t%d %s\n", key, info.Count, info.Size)
	}
}

func getDuplicates(entries map[string]*Info) map[string]*Info {
	duplicates := make(map[string]*Info)
	for key, info := range entries {
		if info.Count >= 2 {
			duplicates[key] = info
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

func detectRuntime(pkgURL string, runtimes map[string]*Info) {
	runtime := ""
	if strings.Contains(pkgURL, "pkg:generic/java/") || strings.Contains(pkgURL, "pkg:maven/") {
		if strings.Contains(pkgURL, "jre") || strings.Contains(pkgURL, "jdk") {
			runtime = "Java"
		}
	} else if strings.Contains(pkgURL, "pkg:python/") {
		if strings.Contains(pkgURL, "cpython") {
			runtime = "Python"
		}
	}
	// TODO: Add more runtime detection rules as needed

	if runtime != "" {
		if runtimes[runtime] == nil {
			runtimes[runtime] = &Info{
				Count: 1,
				Size:  "0", // TODO: implement size determination for runtimes
			}
		} else {
			runtimes[runtime].Count++
		}
	}
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

func generateStats(stats map[string]*Info, title string, filename string, topN int) error {
	if len(stats) == 0 {
		fmt.Fprintf(os.Stderr, "Warn: No duplicates to generate stats for: %s\n", title)
		return nil
	}

	p := plot.New()
	p.Title.Text = title

	type kv struct {
		Key   string
		Value int
	}
	var ss []kv
	for k, v := range stats {
		ss = append(ss, kv{k, v.Count})
	}

	sort.Slice(ss, func(i, j int) bool {
		if ss[i].Value == ss[j].Value {
			return ss[i].Key < ss[j].Key
		}
		return ss[i].Value > ss[j].Value
	})

	if topN > len(ss) {
		topN = len(ss)
	}

	ss = ss[:topN]

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
	bar.Horizontal = true

	p.Add(bar)
	p.NominalY(labels...)

	// Save the plot to a PNG file.
	if err := p.Save(6*vg.Inch, vg.Length(float64(topN)*0.5)*vg.Inch, filename); err != nil {
		return err
	}

	return nil
}
