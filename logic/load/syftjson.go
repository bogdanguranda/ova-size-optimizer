package load

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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
		Path string   `json:"path"`
		Size FileSize `json:"size"`
	} `json:"files"`
}

type FileSize int64

func (s *FileSize) UnmarshalJSON(data []byte) error {
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	switch v := value.(type) {
	case float64:
		*s = FileSize(v)
	case string:
		num, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		*s = FileSize(num)
	default:
		return fmt.Errorf("unexpected type %T for Size", v)
	}

	return nil
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
