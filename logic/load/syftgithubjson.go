package load

import (
	"encoding/json"
	"os"
)

type SyftGithubJSON struct {
	Metadata struct {
		Distro string `json:"syft:distro"`
	} `json:"metadata"`
	Manifests map[string]struct {
		Resolved map[string]struct {
			PackageURL string `json:"package_url"`
		} `json:"resolved"`
	} `json:"manifests"`
}

func LoadGithubJsonSyftFile(file string) (*SyftGithubJSON, error) {
	syftData, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var syftOutput SyftGithubJSON
	if err = json.Unmarshal(syftData, &syftOutput); err != nil {
		return nil, err
	}

	return &syftOutput, nil
}
