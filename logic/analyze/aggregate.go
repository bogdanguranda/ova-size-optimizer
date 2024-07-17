package analyze

import (
	"fmt"

	"github.com/anchore/syft/syft/pkg"
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

	osNameWithVersion := archiveSbom.Artifacts.LinuxDistribution.PrettyName
	if stats.BaseOS[osNameWithVersion] == nil {
		stats.BaseOS[osNameWithVersion] = &Info{
			Count: 1,
			Size:  ConvertSizeBytesToHumanReadableString(GetBaseImageSize(archiveName)),
		}
	} else {
		stats.BaseOS[osNameWithVersion].Count++
	}

	for _, currentPackage := range archiveSbom.Artifacts.Packages.Sorted() {
		if stats.Packages[currentPackage.Name] == nil {
			size := 0
			installedSize := 0

			switch metadata := currentPackage.Metadata.(type) {
			case pkg.ApkDBEntry:
				size = metadata.Size
				installedSize = metadata.InstalledSize
			case pkg.DpkgDBEntry:
				// doesn't have metadata.Size
				installedSize = metadata.InstalledSize
			case pkg.AlpmDBEntry:
				size = metadata.Size
				// doesn't have metadata.InstalledSize
			case pkg.RpmDBEntry:
				size = metadata.Size
				// doesn't have metadata.InstalledSize
			default:
				fmt.Printf("error decoding metadata for package %s of archive %s \n", currentPackage.Name, archiveName)
				continue
			}

			stats.Packages[currentPackage.Name] = &Info{
				Count:         1,
				Size:          ConvertSizeBytesToHumanReadableString(int64(size)),
				InstalledSize: ConvertSizeBytesToHumanReadableString(int64(installedSize)),
			}

		} else {
			stats.Packages[currentPackage.Name].Count++
		}

		// TODO: fix this
		// DetectRuntime(packageInfo[packageName], stats.Runtimes, archiveName)
	}

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

	return stats, nil
}
