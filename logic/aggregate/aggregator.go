package aggregate

import (
	"ova-size-optimizer/logic/load"
)

type Info struct {
	Count         int
	Size          string
	InstalledSize string
}

type Stats struct {
	BaseOS   map[string]*Info
	Packages map[string]*Info
	Runtimes map[string]*Info
}

func NewStats() *Stats {
	return &Stats{
		BaseOS:   make(map[string]*Info),
		Packages: make(map[string]*Info),
		Runtimes: make(map[string]*Info),
	}
}

func ProcessData(stats *Stats, syftGithubJSONOutput *load.SyftGithubJSON, syftJSONOutput *load.SyftJSON, diveOutput *load.DiveJSON) {
	osNameWithVersion := DetectOSNameWithVersion(syftGithubJSONOutput.Metadata.Distro)
	if stats.BaseOS[osNameWithVersion] == nil {
		stats.BaseOS[osNameWithVersion] = &Info{
			Count: 1,
			Size:  ConvertSizeBytesToHumanReadableString(diveOutput.Layers[0].SizeBytes),
		}
	} else {
		stats.BaseOS[osNameWithVersion].Count++
	}

	packageInfo := make(map[string]load.Metadata)
	for _, packageArtifact := range syftJSONOutput.Artifacts {
		packageName := DetectPackageName(packageArtifact.PUrl)
		packageInfo[packageName] = packageArtifact.PackageMetadata
	}

	for _, manifest := range syftGithubJSONOutput.Manifests {
		for _, pkg := range manifest.Resolved {
			packageName := DetectPackageName(pkg.PackageURL)
			if stats.Packages[packageName] == nil {
				stats.Packages[packageName] = &Info{
					Count:         1,
					Size:          ConvertSizeBytesToHumanReadableString(packageInfo[packageName].Size),
					InstalledSize: ConvertSizeBytesToHumanReadableString(packageInfo[packageName].InstalledSize),
				}
			} else {
				stats.Packages[packageName].Count++
			}

			DetectRuntime(pkg.PackageURL, stats.Runtimes)
		}
	}
}
