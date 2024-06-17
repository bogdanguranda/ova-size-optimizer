package aggregate

import (
	"ova-size-optimizer/logic/load"
)

type Info struct {
	Count int
	Size  string
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

func ProcessData(stats *Stats, syftOutput *load.SyftJSON, diveOutput *load.DiveJSON) {
	osNameWithVersion := DetectOSNameWithVersion(syftOutput.Metadata.Distro)
	if stats.BaseOS[osNameWithVersion] == nil {
		stats.BaseOS[osNameWithVersion] = &Info{
			Count: 1,
			Size:  ConvertSizeBytesToHumanReadableString(diveOutput.Layers[0].SizeBytes),
		}
	} else {
		stats.BaseOS[osNameWithVersion].Count++
	}

	for _, manifest := range syftOutput.Manifests {
		for _, pkg := range manifest.Resolved {
			packageName := DetectPackageName(pkg.PackageURL)
			if stats.Packages[packageName] == nil {
				stats.Packages[packageName] = &Info{
					Count: 1,
					Size:  "0", // TODO: implement size determination for packages
				}
			} else {
				stats.Packages[packageName].Count++
			}

			DetectRuntime(pkg.PackageURL, stats.Runtimes)
		}
	}
}
