package aggregate

import (
	"ova-size-optimizer/logic/load"
	"strings"
)

type Info struct {
	Count         int
	Size          string
	InstalledSize string
}

type Stats struct {
	BaseOS   map[string]*Info
	Packages map[string]*Info
	Runtimes map[string]map[string]*Info //map["imageFile"]["runtime"]*Info
}

func NewStats() *Stats {
	return &Stats{
		BaseOS:   make(map[string]*Info),
		Packages: make(map[string]*Info),
		Runtimes: make(map[string]map[string]*Info),
	}
}

func ProcessData(stats *Stats, syftGithubJSONOutput *load.SyftGithubJSON, syftJSONOutput *load.SyftJSON, fileName, individualArchivePathDir string) {

	// full name is <prefix>_<image name>.json
	// eg. syft_json_cp-schema-registry
	// we want to get the image name to reference the tar archive
	tarImageName := strings.Split(strings.Split(fileName, "syft_json_")[1], ".json")[0] + ".tar"

	// construct the full path to hte individual tar archive
	tarPath := individualArchivePathDir + tarImageName
	osNameWithVersion := DetectOSNameWithVersion(syftGithubJSONOutput.Metadata.Distro)
	if stats.BaseOS[osNameWithVersion] == nil {
		stats.BaseOS[osNameWithVersion] = &Info{
			Count: 1,
			Size:  ConvertSizeBytesToHumanReadableString(GetBaseImageSize(tarPath)),
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

			DetectRuntime(packageInfo[packageName], stats.Runtimes, fileName)
		}
	}
}
