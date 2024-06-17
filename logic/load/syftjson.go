package load

import (
	"encoding/json"
	"os"
)

type SyftJSON struct {
	Artifacts []struct {
		PUrl            string   `json:"purl"`
		PackageMetadata Metadata `json:"metadata"`
	} `json:"artifacts"`
}

type Metadata struct {
	PackageName   string `json:"package"`
	OriginPackage string `json:"originPackage"`
	Size          int64  `json:"size"`
	InstalledSize int64  `json:"installedSize"`
	Files         []struct {
		Path string `json:"path"`
		Size int    `json:"size"`
	} `json:"files"`
}

func LoadJsonSyftFile(file string) (*SyftJSON, error) {
	syftData, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var syftOutput SyftJSON
	if err = json.Unmarshal(syftData, &syftOutput); err != nil {
		return nil, err
	}

	return &syftOutput, nil
}
