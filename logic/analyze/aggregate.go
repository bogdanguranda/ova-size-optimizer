package analyze

import (
	"github.com/anchore/syft/syft/sbom"
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

func AggregateData(individualArchivesPathDir, archiveName string, archiveSbom *sbom.SBOM) (*Stats, error) {
	stats := NewStats()
	stats.Runtimes[archiveName] = make(map[string]*Info)

	// TODO: fix this; was: syftGithubJSONOutput.Metadata.Distro

	// tarPath := individualArchivesPathDir + "/" + archiveName
	// osNameWithVersion := DetectOSNameWithVersion(syftGithubJSONOutput.Metadata.Distro)
	// if stats.BaseOS[osNameWithVersion] == nil {
	// 	stats.BaseOS[osNameWithVersion] = &Info{
	// 		Count: 1,
	// 		Size:  ConvertSizeBytesToHumanReadableString(GetBaseImageSize(tarPath)),
	// 	}
	// } else {
	// 	stats.BaseOS[osNameWithVersion].Count++
	// }

	// TODO: fix this; was: syftJSONOutput.Artifacts

	// packageInfo := make(map[string]load.Metadata)
	// for _, packageArtifact := range syftJSONOutput.Artifacts {
	// 	packageName := DetectPackageName(packageArtifact.PUrl)
	// 	packageInfo[packageName] = packageArtifact.PackageMetadata
	// }

	// TODO: fix this; was: syftGithubJSONOutput.Manifests

	// for _, manifest := range syftGithubJSONOutput.Manifests {
	// 	for _, pkg := range manifest.Resolved {
	// 		packageName := DetectPackageName(pkg.PackageURL)
	// 		if stats.Packages[packageName] == nil {
	// 			stats.Packages[packageName] = &Info{
	// 				Count:         1,
	// 				Size:          ConvertSizeBytesToHumanReadableString(packageInfo[packageName].Size),
	// 				InstalledSize: ConvertSizeBytesToHumanReadableString(packageInfo[packageName].InstalledSize),
	// 			}
	// 		} else {
	// 			stats.Packages[packageName].Count++
	// 		}

	// 		DetectRuntime(packageInfo[packageName], stats.Runtimes, archiveName)
	// 	}
	// }

	// TODO: fix this

	// for _, fileName := range syftJSONFilesSlc {
	// 	for k, v := range stats.Runtimes[fileName] {
	// 		sizeInt64, err := strconv.ParseInt(v.Size, 10, 64)
	// 		if err != nil {
	// 			fmt.Printf("Failed to convert string to int64 %v\n", err)
	// 			os.Exit(1)
	// 		}
	// 		v.Size = ConvertSizeBytesToHumanReadableString(sizeInt64)
	// 		stats.Runtimes[fileName][k] = v
	// 	}
	// }

	return nil, nil
}
